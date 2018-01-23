package sns

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
)

type GetAddFriendNewMsgCntReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetAddFriendNewMsgCntRsp struct {
	Cnt int `json:"cnt"`
}

func doGetAddFriendNewMsgCnt(req *GetAddFriendNewMsgCntReq, r *http.Request) (rsp *GetAddFriendNewMsgCntRsp, err error) {

	log.Debugf("uin %d, GetAddFriendNewMsgCntReq %+v", req.Uin, req)

	cnt, err := GetAddFriendNewMsgCnt(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetAddFriendNewMsgCntRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetAddFriendNewMsgCntRsp{cnt}

	log.Debugf("uin %d, GetAddFriendNewMsgCntRsp succ, %+v", req.Uin, rsp)

	return
}

func GetAddFriendNewMsgCnt(uin int64) (cnt int, err error) {

	if uin == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_READ_ADDFRIEND_MSG_ID)
	if err != nil {
		log.Error(err.Error())
		return
	}

	var lastMsgId int

	keyStr := fmt.Sprintf("%d", uin)
	valStr, err := app.Get(keyStr)
	if err != nil {

		//如果KEY不存在 则认为lastMsgId为0
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				valStr = "0"
			} else {
				return
			}
		} else {
			return
		}
	}

	//非法字符 lastTs = 0
	lastMsgId, err1 := strconv.Atoi(valStr)
	if err1 != nil {
		log.Error(err1.Error())
		lastMsgId = 0
	}

	inst2 := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst2 == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select count(msgId) from addFriendMsg where toUin = %d and msgId > %d`, uin, lastMsgId)
	rows, err := inst2.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&cnt)
	}

	return
}
