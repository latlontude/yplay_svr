package sns

import (
	"common/constant"
	//"common/env"
	"api/im"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type AddFriendReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	ToUin   int64 `schema:"toUin"`
	SrcType int   `schema:"srcType"`        //好友来源  9:app_commend
}

type AddFriendRsp struct {
	MsgId int64 `json:"msgId"`
}

func doAddFriend(req *AddFriendReq, r *http.Request) (rsp *AddFriendRsp, err error) {

	log.Errorf("uin %d, AddFriendReq %+v", req.Uin, req)

	msgId, err := AddFriend(req.Uin, req.SrcType, req.ToUin)
	if err != nil {
		log.Errorf("uin %d, AddFriendRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddFriendRsp{msgId}

	log.Errorf("uin %d, AddFriendRsp succ, %+v", req.Uin, rsp)

	//如果申请加好友，就不在共同好友出现
	go Fix2DegreeResult(req.Uin, req.ToUin)

	return
}

func AddFriend(uin int64, srcType int, toUin int64) (msgId int64, err error) {

	if uin == 0 || toUin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	if uin == toUin {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "add self friend")
		log.Error(err.Error())
		return
	}

	isFriend, err := st.CheckIsMyFriend(uin, toUin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if isFriend > 0 {
		log.Errorf("uin %d, toUin %d, already friend!", uin, toUin)
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select msgId from addFriendMsg where fromUin = %d and toUin = %d and status = %d`, uin, toUin, constant.ENUM_ADD_FRIEND_STATUS_INIT)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		rows.Scan(&msgId)
		find = true
	}

	//如果已经发送过请求，则直接返回上次发送的msgid
	if find {
		return
	}

	yes, err := IsInUserBlackList(toUin, uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	sql = fmt.Sprintf(`insert into addFriendMsg values(%d, %d, %d, %d, %d, %d, %d)`, 0, uin, toUin, srcType, constant.ENUM_ADD_FRIEND_STATUS_INIT, ts, 0)

	if yes {
		sql = fmt.Sprintf(`insert into addFriendMsg values(%d, %d, %d, %d, %d, %d, %d)`, 0, uin, toUin, srcType, constant.ENUM_ADD_FRIEND_STATUS_IGNORE, ts, ts)
	}

	res, err := inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	msgId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	if !yes {
		go SendAddFriendPush(uin, toUin)
	}

	return
}

//如果申请了加好友 就不在好友的好友列表展现
func Fix2DegreeResult(uin1, uin2 int64) (err error) {

	if uin1 == 0 || uin2 == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_2DEGREE_FRIENDS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr1 := fmt.Sprintf("%d", uin1)
	keyStr2 := fmt.Sprintf("%d", uin2)

	app.ZRem(keyStr1, keyStr2)
	app.ZRem(keyStr2, keyStr1)

	return

}

func SendAddFriendPush(uin int64, toUin int64) (err error) {

	log.Debugf("SendAddFriendPush (%d->%d) begin", uin, toUin)

	if uin == 0 || toUin == 0 || uin == toUin {
		return
	}

	// ui, err := st.GetUserProfileInfo(uin)
	// if err != nil{
	// 	log.Error(err.Error())
	// 	return
	// }

	// _, err = push.WnsPush(toUin, ui.NickName, "同学～加个好友呗(*/ω＼*)", "this is ext")
	// if err != nil{
	// 	log.Error(err.Error())
	// 	return
	// }

	im.SendAddFriendMsg(uin, toUin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Debugf("SendAddFriendPush succ (%d->%d)", uin, toUin)

	return
}

func IsInUserBlackList(uin, uid int64) (yes bool, err error) {
	log.Debugf("start IsInUserBlackList uin:%d, uid:%d", uin, uid)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select id from pullBlackUser where uin = %d and uid = %d and status = 0`, uin, uid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		rows.Scan(&id)
		yes = true
	}

	log.Debugf("end IsInUserBlackList yes:%t", yes)
	return
}
