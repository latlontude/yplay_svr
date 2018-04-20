package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"time"
)

type ReportUserReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Uid   int64  `schema:"uid"`
	Type  int    `schema:"type"` //举报类型
	Desc  string `schema:"desc"` // 举报文本
}

type ReportUserRsp struct {
}

func doReportUser(req *ReportUserReq, r *http.Request) (rsp *ReportUserRsp, err error) {

	log.Debugf("uin %d, doReportUser %+v", req.Uin, req)

	err = ReportUser(req.Uin, req.Uid, req.Type, req.Desc)
	if err != nil {
		log.Errorf("uin %d, ReportUserRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ReportUserRsp{}

	log.Debugf("uin %d, ReportUserRsp succ, %+v", req.Uin, rsp)

	return
}

func ReportUser(uin, uid int64, typ int, desc string) (err error) {
	log.Debugf("start ReportUser uin:%d uid:%d typ:%d desc:%s", uin, uid, typ, desc)

	if typ < 1 || typ > 5 {
		err = rest.NewAPIError(constant.E_INVALID_REPORT_TYPE, "report type invalid")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	stmt, err := inst.Prepare(`replace into reportUser values(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, uid, typ, desc, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	log.Debugf("end ReportUser")
	return
}
