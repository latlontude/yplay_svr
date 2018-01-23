package submit

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type SubmitRejectReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SubmitId int    `schema:"submitId"`
	User     int64  `schema:"user"`
	Desc     string `schema:"desc"`
}

type SubmitRejectRsp struct {
}

func doSubmitReject(req *SubmitRejectReq, r *http.Request) (rsp *SubmitRejectRsp, err error) {

	log.Errorf("uin %d, SubmitRejectReq %+v", req.Uin, req)

	err = SubmitReject(req.Uin, req.SubmitId, req.User, req.Desc)
	if err != nil {
		log.Errorf("uin %d, SubmitRejectRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitRejectRsp{}

	log.Errorf("uin %d, SubmitRejectRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitReject(uin int64, submitId int, user int64, desc string) (err error) {

	if submitId == 0 || user == 0 || len(desc) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select qtext from submitQuestions where id = %d and uin = %d and status = %d`, submitId, user, 0)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		var tmp string
		rows.Scan(&tmp)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "res not found")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()

	stmt, err := inst.Prepare(`update submitQuestions set mts = ?, status = ?, descr = ? where id = ?`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(ts, 2, desc, submitId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
