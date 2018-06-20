package answer

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"time"
)

type PostAnswerReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid           int    `schema:"qid"`
	AnswerContent string `schema:"answerContent"`
	AnswerImgUrls string `schema:"answerImgUrls"`
}

type PostAnswerRsp struct {
	Qid      int `json:"qid"`
	AnswerId int `json:"answerId"`
}

func doPostAnswer(req *PostAnswerReq, r *http.Request) (rsp *PostAnswerRsp, err error) {

	log.Debugf("uin %d, PostAnswerReq %+v", req.Uin, req)

	answerId, err := PostAnswer(req.Uin, req.Qid, req.AnswerContent, req.AnswerImgUrls)

	if err != nil {
		log.Errorf("uin %d, PostAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PostAnswerRsp{req.Qid, int(answerId)}

	log.Debugf("uin %d, PostQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func PostAnswer(uin int64, qid int, answerContent, answerImgUrls string) (answerId int64, err error) {

	if len(answerContent) == 0 && len(answerImgUrls) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into v2answers(answerId, qid, answerContent, answerImgUrls, ownerUid, answerStatus, answerTs) 
		values(?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	res, err := stmt.Exec(0, qid, answerContent, answerImgUrls, uin, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	answerId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
