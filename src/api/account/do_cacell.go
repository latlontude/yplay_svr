package account

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/rest"
	//"common/myredis"
	"common/myredis"
	"fmt"
	"net/http"
	"svr/st"
)

type CanCellReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	UserName string `schema:"userName"`
}

type CanCellRsp struct {

}

func doCancell3(req *CanCellReq, r *http.Request) (rsp *CanCellRsp, err error) {

	log.Errorf("uin %d, CancellReq %+v", req.Uin, req)

	err = Cancell(req.UserName)
	if err != nil {
		log.Errorf("uin %d, CancellRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &CanCellRsp{}

	log.Errorf("uin %d, CancellRsp succ, %+v", req.Uin, rsp)

	return
}

func Cancell(userName string) (err error) {

	if len(userName) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, phone from profiles where userName = "%s"`, userName)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false
	var phone string
	var uin int64

	for rows.Next() {
		rows.Scan(&uin, &phone)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "permission denied")
		log.Error(err.Error())
		return
	}

	friends, err := st.GetMyFriendUins(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//注销账号
	sqls := make([]string, 0)
	sqls = append(sqls, fmt.Sprintf(`delete from profiles where uin = %d`, uin))

	//好友信息删掉
	sqls = append(sqls, fmt.Sprintf(`delete from friends where friendUin = %d`, uin))
	sqls = append(sqls, fmt.Sprintf(`delete from friends where uin = %d`, uin))

	//通讯录的状态
	sqls = append(sqls, fmt.Sprintf(`update addrBook set friendUin = 0 where friendPhone = "%s"`, phone))

	//解除好友之后把之前的添加好友消息都清理掉
	sqls = append(sqls, fmt.Sprintf(`delete from addFriendMsg where fromUin = %d`, uin))
	sqls = append(sqls, fmt.Sprintf(`delete from addFriendMsg where toUin = %d`, uin))

	//短信邀请信息删掉
	sqls = append(sqls, fmt.Sprintf(`delete from inviteFriendSms where phone = "%s"`, phone))
	sqls = append(sqls, fmt.Sprintf(`delete from inviteFriendSms where uin = %d`, uin))
	sqls = append(sqls, fmt.Sprintf(`delete from invitecode where phone = "%s"`, phone))

	sqls = append(sqls, fmt.Sprintf(`delete from freezingStatus where uin = "%s"`, uin))
	sqls = append(sqls, fmt.Sprintf(`delete from profileModRecords where uin = "%s"`, uin))

	//统计信息删掉
	sqls = append(sqls, fmt.Sprintf(`delete from userStat where uin = %d`, uin))

	for _, fUin := range friends {
		sqls = append(sqls, fmt.Sprintf(`update userStat set statValue = statValue -1 where uin = %d and statField = %d and statValue > 0`, fUin, constant.ENUM_USER_STAT_FRIEND_CNT))
	}

	err = mydb.Exec(inst, sqls)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_TOKEN)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = app.Del(fmt.Sprintf("%d", uin))
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
