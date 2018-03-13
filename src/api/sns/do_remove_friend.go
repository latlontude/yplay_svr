package sns

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/rest"
	//"common/myredis"
	"api/im"
	"fmt"
	"net/http"
	"time"
)

type RemoveFriendReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	ToUin int64 `schema:"toUin"`
}

type RemoveFriendRsp struct {
}

func doRemoveFriend(req *RemoveFriendReq, r *http.Request) (rsp *RemoveFriendRsp, err error) {

	log.Errorf("uin %d, RemoveFriendReq %+v", req.Uin, req)

	err = RemoveFriend(req.Uin, req.ToUin)
	if err != nil {
		log.Errorf("uin %d, RemoveFriendRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &RemoveFriendRsp{}

	log.Errorf("uin %d, RemoveFriendRsp succ, %+v", req.Uin, rsp)

	return
}

func RemoveFriend(uin, friendUin int64) (err error) {

	if uin == 0 || friendUin == 0 || uin == friendUin {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select status from friends where uin = %d and friendUin = %d`, uin, friendUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false
	var status int

	for rows.Next() {
		rows.Scan(&status)
		find = true
	}

	//已经不是好友，直接返回成功
	if !find {
		//err = rest.NewAPIError(constant.E_PERMI_DENY, "permission denied")
		//log.Error(err.Error())
		return
	}

	ts := time.Now().Unix()
	ms := time.Now().UnixNano() / 1000000

	//添加到我的好友列表中
	sqls := make([]string, 0)
	sqls = append(sqls, fmt.Sprintf(`delete from friends where uin = %d and friendUin = %d`, uin, friendUin))
	sqls = append(sqls, fmt.Sprintf(`delete from friends where uin = %d and friendUin = %d`, friendUin, uin))
	//更新我的好友列表的版本号
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, friendUin, ms, ts, ms, ts))
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, uin, ms, ts, ms, ts))

	//解除好友之后把之前的添加好友消息都清理掉
	sqls = append(sqls, fmt.Sprintf(`delete from addFriendMsg where fromUin = %d and toUin = %d`, uin, friendUin))
	sqls = append(sqls, fmt.Sprintf(`delete from addFriendMsg where fromUin = %d and toUin = %d`, friendUin, uin))

	//解除好友之后，好友计数要减少
	sqls = append(sqls, fmt.Sprintf(`update userStat set statValue = statValue -1 where uin = %d and statField = %d and statValue > 0`, uin, constant.ENUM_USER_STAT_FRIEND_CNT))
	sqls = append(sqls, fmt.Sprintf(`update userStat set statValue = statValue -1 where uin = %d and statField = %d and statValue > 0`, friendUin, constant.ENUM_USER_STAT_FRIEND_CNT))

	err = mydb.Exec(inst, sqls)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//更新好友的好友关系
	//.......
	go im.SendRemoveFriendMsg(uin, friendUin)

	return
}
