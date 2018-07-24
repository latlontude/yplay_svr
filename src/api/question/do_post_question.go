package question

import (
	"api/elastSearch"
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"strings"
	"time"
)

type PostQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId     int    `schema:"boardId"`
	QTitle      string `schema:"qTitle"`
	QContent    string `schema:"qContent"`
	QImgUrls    string `schema:"qImgUrls"`
	IsAnonymous bool   `schema:"isAnonymous"`
}

type PostQuestionRsp struct {
	Qid int `json:"qid"`
}

func doPostQuestion(req *PostQuestionReq, r *http.Request) (rsp *PostQuestionRsp, err error) {

	log.Debugf("uin %d, PostQuestionReq %+v", req.Uin, req)

	//去除首位空白字符
	qid, err := PostQuestion(req.Uin, req.BoardId, req.QTitle, strings.Trim(req.QContent, " \n\t"), req.QImgUrls, req.IsAnonymous)

	if err != nil {
		log.Errorf("uin %d, PostQuestion error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PostQuestionRsp{int(qid)}

	log.Debugf("uin %d, PostQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func PostQuestion(uin int64, boardId int, title, content, imgUrls string, isAnonymous bool) (qid int64, err error) {

	if boardId == 0 {
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

	//v2question表多加了一个字段  (同问sameAskUid)
	stmt, err := inst.Prepare(`insert into v2questions(qid, boardId, ownerUid, qTitle, qContent, qImgUrls, isAnonymous, qStatus, createTs, modTs,sameAskUid) 
		values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认

	res, err := stmt.Exec(0, boardId, uin, title, content, imgUrls, isAnonymous, status, ts, 0, "")
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	qid, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	err1 := elastSearch.AddQstToEs(boardId, int(qid), content)
	if err1 != nil {
		log.Debugf("es error")
	}

	return
}
