package question

import (
	"api/elastSearch"
	"api/v2push"
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
	QType       int    `schema:"qType"`
	IsAnonymous bool   `schema:"isAnonymous"`
	Ext         string `schema:"ext"`

	Longitude float64 `schema:"longitude"`
	Latitude  float64 `schema:"latitude"`
	PoiTag    string  `schema:"poiTag"`
}

type PostQuestionRsp struct {
	Qid int `json:"qid"`
}

func doPostQuestion(req *PostQuestionReq, r *http.Request) (rsp *PostQuestionRsp, err error) {

	log.Debugf("uin %d, PostQuestionReq %+v", req.Uin, req)

	//去除首位空白字符
	qid, err := PostQuestion(req.Uin, req.BoardId, req.QTitle, strings.Trim(req.QContent, " \n\t"), req.QImgUrls,
		req.QType, req.IsAnonymous, req.Ext, req.Longitude, req.Latitude, req.PoiTag)

	if err != nil {
		log.Errorf("uin %d, PostQuestion error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PostQuestionRsp{int(qid)}

	log.Debugf("uin %d, PostQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func PostQuestion(uin int64, boardId int, title, content, imgUrls string, qType int, isAnonymous bool, ext string,
	lng float64, lat float64, poi string) (qid int64, err error) {

	log.Debugf("post questions")
	if boardId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	if len(content) == 0 && len(imgUrls) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "content and img empty")
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
	stmt, err := inst.Prepare(`insert into v2questions(qid, boardId, ownerUid, qTitle, qContent, qImgUrls, qType,isAnonymous, qStatus, createTs,
		modTs,sameAskUid,ext,longitude, latitude, poiTag)	values(?,?,?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ,?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认

	res, err := stmt.Exec(0, boardId, uin, title, content, imgUrls, qType, isAnonymous, status, ts, 0, "", ext, lng, lat, poi)
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

	var qstInter interface{}

	//if uin == 103096 {
	go v2push.SendAtPush(uin, 1, int(qid), qstInter, ext)
	//}

	//把问题添加到mgo中
	AddQuestionToMgo(uin, int(qid), lat, lng, poi, ts)

	return
}
