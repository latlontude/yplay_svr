package submit

import (
	"api/geneqids"
	"api/im"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/cache"
	"time"
)

type BatchSubmitApproveReq struct {
	MinId int `schema:"minId"`
	MaxId int `schema:"maxId"`
	typ   int `schema:"type"`
}

type BatchSubmitApproveRsp struct {
	Qids []int64 `schema:"qids"`
}

type SubmitApproveReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SubmitId int   `schema:"submitId"`
	User     int64 `schema:"user"`
	typ      int   `schema:"type"` //type : 0, 投给同校，1 投给同校同年级
}

type SubmitApproveRsp struct {
	QId int `json:"qid"`
}

func doBatchSubmitApprove(req *BatchSubmitApproveReq, r *http.Request) (rsp *BatchSubmitApproveRsp, err error) {

	log.Errorf("doBatchSubmitApproveReq minId:%d, maxId:%d,typ:%d", req.MinId, req.MaxId, req.typ)

	qids, err := BatchSubmit(req.MinId, req.MaxId, req.typ)
	if err != nil {
		log.Errorf("BatchSubmitApproveRsp error, %s", err.Error())
		return
	}

	rsp = &BatchSubmitApproveRsp{qids}

	log.Errorf("doBatchSubmitApproveRsp succ, %+v", rsp)

	return
}

func doSubmitApprove(req *SubmitApproveReq, r *http.Request) (rsp *SubmitApproveRsp, err error) {

	log.Errorf("uin %d, SubmitApproveReq %+v", req.Uin, req)

	qid, err := SubmitApprove(req.Uin, req.User, req.SubmitId, req.typ)
	if err != nil {
		log.Errorf("uin %d, SubmitApproveRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitApproveRsp{int(qid)}

	log.Errorf("uin %d, SubmitApproveRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitApprove(uin, user int64, submitId, typ int) (qid int64, err error) {
	log.Debugf("start SubmitApprove uin:%d, user:%d, submitId:%d, typ:%d", uin, user, submitId, typ)

	if submitId == 0 || user == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select qtext, qiconId from submitQuestions where id = %d and uin = %d and status != %d`, submitId, user, 1)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	qtext := ""
	qiconId := 0
	find := false
	for rows.Next() {
		rows.Scan(&qtext, &qiconId)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "res not found")
		return
	}

	//插入到题目数据库
	stmt, err := inst.Prepare(`insert into questions2(qid, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, delivery, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	optionGender := 0
	replyGender := 0
	schoolType := 0
	status := 0
	dataSrc := 2    //投稿题库
	delivery := typ //同校同年级可见,同校可见

	tagId := 0
	tagName := ""

	subTagId1 := 0
	subTagId2 := 0
	subTagId3 := 0
	subTagName1 := ""
	subTagName2 := ""
	subTagName3 := ""

	qiconUrl := fmt.Sprintf("%d.png", qiconId)

	res, err := stmt.Exec(0, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, delivery, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//获取到新增的题目ID
	qid, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//更新审核题目信息和状态
	sql = fmt.Sprintf(`update submitQuestions set qid = %d, mts = %d, status = %d where id = %d`, qid, ts, 1, submitId)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//插入到临时数据库
	//go InsertApprovedQuestion2Tmp(int(qid), qtext, qiconId)

	//缓存的question要增加
	go cache.AddCacheQuestions(int(qid))
	//发送IM消息
	go im.SendSubmitQustionApprovedMsg(user)

	//审核通过的题目立即插入到用户的未答题的列表里面
	go geneqids.ApproveQuestionUpdate(user, int(qid), typ)

	log.Debugf("end SubmitApprove uin:%d, user:%d, submitId:%d, typ:%d", uin, user, submitId, typ)
	return
}

func BatchSubmit(minId, maxId, typ int) (qids []int64, err error) {

	if minId == 0 || maxId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}
	id := minId

	for {
		if id > maxId {
			break
		}

		qid, err := BatchSubmitApprove(id, typ)
		if err != nil {
			log.Errorf(err.Error())
			id++
			continue
		} else {
			qids = append(qids, qid)
		}
		id++
	}

	return
}

func BatchSubmitApprove(submitId, typ int) (qid int64, err error) {

	log.Debugf("start BatchSubmitApprove submitId:%d", submitId)

	if submitId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select qtext, qiconId, uin from submitQuestions where id = %d  and status != %d`, submitId, 1)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	qtext := ""
	qiconId := 0
	var user int64
	find := false
	for rows.Next() {
		rows.Scan(&qtext, &qiconId, &user)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "res not found")
		return
	}

	log.Debugf("id:%d, user:%d", submitId, user)

	//插入到题目数据库
	stmt, err := inst.Prepare(`insert into questions2(qid, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, delivery, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	optionGender := 0
	replyGender := 0
	schoolType := 0
	status := 0
	dataSrc := 2    //投稿题库
	delivery := typ //同校同年级可见,同校可见
	tagId := 0
	tagName := ""

	subTagId1 := 0
	subTagId2 := 0
	subTagId3 := 0
	subTagName1 := ""
	subTagName2 := ""
	subTagName3 := ""

	qiconUrl := fmt.Sprintf("%d.png", qiconId)

	res, err := stmt.Exec(0, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, delivery, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//获取到新增的题目ID
	qid, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//更新审核题目信息和状态
	sql = fmt.Sprintf(`update submitQuestions set qid = %d, mts = %d, status = %d where id = %d`, qid, ts, 1, submitId)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//插入到临时数据库
	//go InsertApprovedQuestion2Tmp(int(qid), qtext, qiconId)

	//缓存的question要增加
	go cache.AddCacheQuestions(int(qid))
	//发送IM消息
	go im.SendSubmitQustionApprovedMsg(user)

	//审核通过的题目立即插入到用户的未答题的列表里面
	go geneqids.ApproveQuestionUpdate(user, int(qid), typ)

	log.Debugf("end BatchSubmitApprove submitId:%d", submitId)
	return
}

/*
func InsertApprovedQuestion2Tmp(qid int, qtext string, qiconId int) (err error) {

	if qid == 0 || len(qtext) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	//插入到题目数据库
	stmt, err := inst.Prepare(`insert into questionsTmp(qid, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	optionGender := 0
	replyGender := 0
	schoolType := 0
	status := 0
	dataSrc := 2 //投稿题库

	tagId := 0
	subTagId1 := 0
	subTagId2 := 0
	subTagId3 := 0

	tagName := ""
	subTagName1 := ""
	subTagName2 := ""
	subTagName3 := ""

	qiconUrl := fmt.Sprintf("%d.png", qiconId)

	_, err = stmt.Exec(qid, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}
*/
