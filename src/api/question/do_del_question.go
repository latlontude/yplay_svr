package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid int `schema:"qid"`
}

type DelQuestionRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelQuestion(req *DelQuestionReq, r *http.Request) (rsp *DelQuestionRsp, err error) {

	log.Debugf("uin %d, DelQuestionReq %+v", req.Uin, req)

	code, err := DelQuestion(req.Uin, req.Qid)

	if err != nil {
		log.Errorf("uin %d, DelQuestion error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelQuestionRsp{code}

	log.Debugf("uin %d, DelQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func DelQuestion(uin int64, qid int) (code int, err error) {
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

	uids, err := getDelQidPermitOperators(qid)
	if err != nil {
		log.Error(err)
		return
	}

	permit := false
	for _, uid := range uids {
		if uid == uin {
			permit = true
		}
	}

	if !permit {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "no permissions")
		log.Errorf("uin:%d has no permissions to delete question:%d", qid)
		return
	}

	sql := fmt.Sprintf(`update v2questions set qStatus = 1 where qid = %d`, qid)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	code = 0

	log.Debugf("end DelQuestion uin = %d qid = %d code = %d", uin, qid, code)
	return
}

func getDelQidPermitOperators(qid int) (operators []int64, err error) {

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
	var owner int64
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
	return
}
