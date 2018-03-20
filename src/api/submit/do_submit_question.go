package submit

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"svr/cache"
	"time"
)

type SubmitQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	QText   string `schema:"qtext"`
	QIconId int    `schema:"qiconId"`
}

type SubmitQuestionRsp struct {
}

func doSubmitQuestion(req *SubmitQuestionReq, r *http.Request) (rsp *SubmitQuestionRsp, err error) {

	log.Errorf("uin %d, SubmitQuestionReq %+v", req.Uin, req)

	err = SubmitQuestion(req.Uin, req.QIconId, req.QText)
	if err != nil {
		log.Errorf("uin %d, SubmitQuestionRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitQuestionRsp{}

	log.Errorf("uin %d, SubmitQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitQuestion(uin int64, iconId int, qtext string) (err error) {

	if uin == 0 || iconId == 0 || len(qtext) == 0 {
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

	stmt, err := inst.Prepare(`insert ignore into submitQuestionstmp values(?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	qid := 0
	status := 0
	desc := ""

	_, err = stmt.Exec(0, uin, qtext, iconId, qid, status, desc, ts, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
