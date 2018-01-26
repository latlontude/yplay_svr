package vote

import (
	"net/http"
	"svr/st"
        "time"
        "common/constant"
        "common/mydb"
        "common/rest"
)

type SkipQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	QId   int `schema:"qid"`
	Index int `schema:"index"`
}

type SkipQuestionRsp struct {
}

func doSkipQuestion(req *SkipQuestionReq, r *http.Request) (rsp *SkipQuestionRsp, err error) {

	log.Debugf("uin %d, SkipQuestionReq %+v", req.Uin, req)

	//if req.Uin == 100328 || req.Uin == 100446 {
	if true {
		err = SkipQuestionPreGene(req.Uin, req.QId, req.Index)
	} else {
		err = SkipQuestion(req.Uin, req.QId, req.Index)
	}

	if err != nil {
		log.Errorf("uin %d, SkipQuestionRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SkipQuestionRsp{}

	log.Debugf("uin %d, SkipQuestionRsp succ, %+v", req.Uin, rsp)

	ts := time.Now().Unix()
	act := 0
	qid := req.QId
        uin := req.Uin

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into actRecords values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, qid, act, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

func SkipQuestion(uin int64, qid int, index int) (err error) {

	err = st.UpdateVoteProgress2(uin, qid, index)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}

func SkipQuestionPreGene(uin int64, qid int, index int) (err error) {

	err = st.UpdateVoteProgressByPreGene(uin, qid, index)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
