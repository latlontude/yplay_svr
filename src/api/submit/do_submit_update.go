package submit

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/cache"
	"time"
)

type SubmitUpdateReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SubmitId int    `schema:"submitId"`
	QText    string `schema:"qtext"`
	QIconId  int    `schema:"qiconId"`
}

type SubmitUpdateRsp struct {
}

func doSubmitUpdate(req *SubmitUpdateReq, r *http.Request) (rsp *SubmitUpdateRsp, err error) {

	log.Errorf("uin %d, SubmitUpdateReq %+v", req.Uin, req)

	err = SubmitUpdate(req.Uin, req.SubmitId, req.QIconId, req.QText)
	if err != nil {
		log.Errorf("uin %d, SubmitUpdateRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitUpdateRsp{}

	log.Errorf("uin %d, SubmitUpdateRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitUpdate(uin int64, submitId int, iconId int, qtext string) (err error) {

	if uin == 0 || iconId == 0 || len(qtext) == 0 || submitId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	_, ok := cache.QICONS[iconId]

	if !ok {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "iconid not found")
		log.Error(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	sql := fmt.Sprintf(`select qtext from submitQuestions where id = %d and uin = %d and status = %d`, submitId, uin, 2)
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

	stmt, err := inst.Prepare(`update submitQuestions set qiconId = ?, qtext = ?, mts = ?, status = ? where id = ?`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(iconId, qtext, ts, 0, submitId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
