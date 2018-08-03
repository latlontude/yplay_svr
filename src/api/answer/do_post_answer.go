package answer

import (
	"api/elastSearch"
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"strings"
	"svr/st"
	"time"
)

type PostAnswerReq struct {
	Uin           int64  `schema:"uin"`
	Token         string `schema:"token"`
	Ver           int    `schema:"ver"`
	BoardId       int    `schema:"boardId"`
	Qid           int    `schema:"qid"`
	AnswerContent string `schema:"answerContent"`
	AnswerImgUrls string `schema:"answerImgUrls"`
	Ext           string `schema:"ext"`
}

type PostAnswerRsp struct {
	Qid      int `json:"qid"`
	AnswerId int `json:"answerId"`
}

func doPostAnswer(req *PostAnswerReq, r *http.Request) (rsp *PostAnswerRsp, err error) {

	log.Debugf("uin %d, PostAnswerReq %+v", req.Uin, req)

	answerId, err := PostAnswer(req.Uin, req.BoardId, req.Qid, strings.Trim(req.AnswerContent, " \n\t"), req.AnswerImgUrls, req.Ext)

	if err != nil {
		log.Errorf("uin %d, PostAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PostAnswerRsp{req.Qid, int(answerId)}

	log.Debugf("uin %d, PostQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func PostAnswer(uin int64, boardId, qid int, answerContent, answerImgUrls, ext string) (answerId int64, err error) {

	if len(answerContent) == 0 && len(answerImgUrls) == 0 && len(ext) == 0 {
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

	stmt, err := inst.Prepare(`insert into v2answers(answerId, qid, answerContent, answerImgUrls, ownerUid, answerStatus, answerTs,ext) 
		values(?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	res, err := stmt.Exec(0, qid, answerContent, answerImgUrls, uin, status, ts, ext)
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

	err1 := elastSearch.AddAnswerToEs(boardId, qid, int(answerId), answerContent)
	if err1 != nil {
		log.Debugf("add answer to es error")
	}

	var newAnswer st.AnswersInfo
	newAnswer.Qid = qid
	newAnswer.AnswerId = int(answerId)
	newAnswer.AnswerContent = answerContent
	newAnswer.AnswerImgUrls = answerImgUrls
	newAnswer.AnswerTs = int(ts)
	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}
	newAnswer.OwnerInfo = ui

	//给提问者和回答过这道题目的人发送新增回答通知,把回答者uin带过去
	if len(ext) > 0 {
		go v2push.SendAtPush(uin, 2, qid, newAnswer, ext)
	} else {
		go v2push.SendNewAddAnswerPush(uin, qid, newAnswer)
	}

	return
}
