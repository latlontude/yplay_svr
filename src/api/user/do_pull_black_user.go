package user

import (
	"api/sns"
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"time"
)

type PullBlackUserReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Uid   int64  `schema:"uid"`
}

type PullBlackUserRsp struct {
}

func doPullBlackUser(req *PullBlackUserReq, r *http.Request) (rsp *PullBlackUserRsp, err error) {

	log.Debugf("uin %d, doPullBlackUser %+v", req.Uin, req)

	err = PullBlackUser(req.Uin, req.Uid)
	if err != nil {
		log.Errorf("uin %d, PullBlackUserRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PullBlackUserRsp{}

	log.Debugf("uin %d, PullBlackUserRsp succ, %+v", req.Uin, rsp)

	return
}

func PullBlackUser(uin, uid int64) (err error) {
	log.Debugf("start PullBlackUser uin:%d uid:%d ", uin, uid)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	stmt, err := inst.Prepare(`replace into pullBlackUser values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, uid, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	go sns.RemoveFriend(uin, uid)
	log.Debugf("end PullBlackUser")
	return
}
