package sns

import (
	"api/im"
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type AddCustomServiceAccountReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type AddCustomServiceAccountRsp struct {
}

func doAddCustomServiceAccount(req *AddCustomServiceAccountReq, r *http.Request) (rsp *AddCustomServiceAccountRsp, err error) {

	log.Errorf("uin %d, doAddCustomServiceAccountReq %+v", req.Uin, req)

	err = AddCustomServiceAccount(req.Uin)
	if err != nil {
		log.Errorf("uin %d, AddCustomServiceAccount error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddCustomServiceAccountRsp{}

	log.Errorf("uin %d, doAddCustomServiceAccountRsp succ, %+v", req.Uin, rsp)

	return
}

func AddCustomServiceAccount(uin int64) (err error) {

	log.Debugf("start AddCustomServiceAccount uin:%d", uin)

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}
	//检查两人是否已经是好友，如果是则计数不用再增加
	sql := fmt.Sprintf(`select uin from friends where uin = %d and friendUin = %d`, uin, serviceAccountUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)
		find = true
	}

	//如果已经是好友，则直接返回，计数和插入操作免去
	if find {
		log.Debugf("%d has been friend with service account %d", uin, serviceAccountUin)
		return
	}

	//更新状态
	ts := time.Now().Unix()
	ms := time.Now().UnixNano() / 1000000 //毫秒
	//添加到我的好友列表中
	sqls := make([]string, 0)
	sqls = append(sqls, fmt.Sprintf(`insert ignore into friends values(%d, %d, %d, %d)`, serviceAccountUin, uin, 0, ts))
	sqls = append(sqls, fmt.Sprintf(`insert ignore into friends values(%d, %d, %d, %d)`, uin, serviceAccountUin, 0, ts))

	//更新我的好友列表的版本号
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, uin, ms, ts, ms, ts))
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, serviceAccountUin, ms, ts, ms, ts))

	err = mydb.Exec(inst, sqls)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	go IncrFriendCnt(uin, serviceAccountUin)

	//更新好友的好友关系
	im.SendAcceptAddFriendMsg(uin, serviceAccountUin)

	sessionId := ""
	sessionId, err = getSnapSession(uin, serviceAccountUin)
	if err != nil || len(sessionId) == 0 {
		sessionId, err = CreateNewSnapSession(uin, serviceAccountUin)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
	}

	text := "欢迎来到噗噗！如果你在使用中有任何的问题或建议，非常欢迎与我沟通！噗——"
	go im.SendTextMsg(sessionId, text, serviceAccountUin, uin)
	log.Debugf("end AddCustomServiceAccount")

	return
}

func getSnapSession(uin, uid int64) (sessionId string, err error) {
	log.Debugf("start getSnapSession uin:%d uid:%d", uin, uid)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SNAPCHAT_SESSION)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_%d", uin, uid)
	if uin > uid {
		keyStr = fmt.Sprintf("%d_%d", uid, uin)
	}

	val, err := app.Get(keyStr)
	if err != nil {
		log.Error(err.Error())
	}
	sessionId = val

	log.Debugf("end getSnapSession sessionId:%s", sessionId)
	return
}
