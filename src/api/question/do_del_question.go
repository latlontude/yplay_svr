package question

import (
	"api/answer"
	"api/common"
	"api/elastSearch"
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
)

type DelQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid     int    `schema:"qid"`
	Reason  string `schema:"reason"` //删除原因
	Version int    `schema:"version"`
}

type DelQuestionRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelQuestion(req *DelQuestionReq, r *http.Request) (rsp *DelQuestionRsp, err error) {

	log.Debugf("uin %d, DelQuestionReq %+v", req.Uin, req)

	code, err := DelQuestion(req.Uin, req.Qid, req.Reason, req.Version)

	if err != nil {
		log.Errorf("uin %d, DelQuestion error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelQuestionRsp{code}

	log.Debugf("uin %d, DelQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func DelQuestion(uin int64, qid int, reason string, version int) (code int, err error) {
	log.Debugf("start DelQuestion uin = %d qid = %d", uin, qid)

	code = -1

	if uin <= 0 || qid <= 0 {
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

	uids, ownerUid, err := getDelQidPermitOperators(qid)
	if err != nil {
		log.Error(err)
		return
	}

	permit := false
	isMyself := false
	for _, uid := range uids {
		if uid == uin {
			permit = true
		}
		if ownerUid == uin {
			isMyself = true
		}
	}

	if !permit {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "no permissions")
		log.Errorf("uin:%d has no permissions to delete question:%d", uin, qid)
		return
	}

	sql := fmt.Sprintf(`update v2questions set qStatus = 1 where qid = %d`, qid)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//从elastSearch删掉提问
	err2 := elastSearch.DelQstToEs(qid)
	if err2 != nil {
		log.Errorf(err2.Error())
	}

	//删除该问题的所有回答
	sql = fmt.Sprintf(`select answerId from v2answers where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var answerId int
		rows.Scan(&answerId)
		//var pupuUin int64 = 100001
		_, err2 := answer.DelAnswer(uin, qid, answerId, "")
		//err2 := elastSearch.DelAnswerFromEs(answerId)
		if err2 != nil {
			log.Errorf(err2.Error())
		}
	}

	//不是我自己删的  发推送
	if !isMyself {
		question, _ := common.GetV2Question(qid, version)
		data, err1 := json.Marshal(&question)
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}
		dataStr := string(data)
		v2push.SendBeDeletePush(uin, ownerUid, reason, 1, dataStr)
	}

	code = 0

	log.Debugf("end DelQuestion uin = %d qid = %d code = %d", uin, qid, code)
	return
}

func getDelQidPermitOperators(qid int) (operators []int64, owner int64, err error) {

	log.Debugf("start getDelQidPermitOperators qid:%d", qid)
	if qid == 0 {
		log.Errorf("qid is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	boardId := 0
	sql := fmt.Sprintf(`select boardId from v2questions where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	for rows.Next() {
		rows.Scan(&boardId)
	}

	//本版块的墙主有权限删除本版块的问题
	var manager int64
	sql = fmt.Sprintf(`select ownerUid from v2boards where boardId = %d`, boardId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&manager)
	}

	if manager != 0 {
		operators = append(operators, manager)
	}

	//这道题目的提问者有权限删除本问题
	sql = fmt.Sprintf(`select ownerUid from v2questions where qid = %d`, qid)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&owner)
	}

	if owner != 0 {
		operators = append(operators, owner)
	}

	//pupu客服有权删除
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	operators = append(operators, serviceAccountUin)
	log.Debugf("end getDelQidPermitOperators qid:%d uids:%+v", qid, operators)
	return
}
