package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type RemoveUserFromMyBlacklistReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Uid   int64  `schema:"uid"`
}

type RemoveUserFromMyBlacklistRsp struct {
}

func doRemoveUserFromMyBlacklist(req *RemoveUserFromMyBlacklistReq, r *http.Request) (rsp *RemoveUserFromMyBlacklistRsp, err error) {

	log.Debugf("uin %d, doRemoveUserFromMyBlacklist uid:%d", req.Uin, req.Uid)

	err = RemoveUserFromMyBlacklist(req.Uin, req.Uid)
	if err != nil {
		log.Errorf("uin %d, RemoveUserFromMyBlacklistRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &RemoveUserFromMyBlacklistRsp{}

	log.Debugf("uin %d, GetMyBlacklistRsp succ, %+v", req.Uin, rsp)

	return
}

func RemoveUserFromMyBlacklist(uin, uid int64) (err error) {
	log.Debugf("start RemoveUserFromMyBlacklist uin:%d ", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`delete from pullBlackUser where uin = %d and uid = %d`, uin, uid)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end RemoveUserFromMyBlacklist")
	return
}
