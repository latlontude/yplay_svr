package answer

import (
	"api/common"
	"api/elastSearch"
	"api/experience"
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
)

type DelAnswerReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid      int    `schema:"qid"`
	AnswerId int    `schema:"answerId"`
	Reason   string `schema:"reason"`
}

type DelAnswerRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelAnswer(req *DelAnswerReq, r *http.Request) (rsp *DelAnswerRsp, err error) {

	log.Debugf("uin %d, DelAnswerReq %+v", req.Uin, req)

	code, err := DelAnswer(req.Uin, req.Qid, req.AnswerId, req.Reason)

	if err != nil {
		log.Errorf("uin %d, DelAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelAnswerRsp{code}

	log.Debugf("uin %d, DelAnswerRsp succ, %+v", req.Uin, rsp)

	return
}

func DelAnswer(uin int64, qid, answerId int, reason string) (code int, err error) {
	log.Debugf("start DelAnswer uin = %d qid = %d answerId = %d", uin, qid, answerId)

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

	uids, ownerUid, qidOwner, err := getDelAnswerPermitOperators(answerId, qid)
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

		if uin == ownerUid {
			isMyself = true
		}
	}

	if !permit {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "no permissions")
		log.Errorf("uin:%d has no permissions to delete answer:%d", uin, answerId)
		return
	}

	sql := fmt.Sprintf(`update v2answers set answerStatus = 1 where qid = %d and answerId = %d`, qid, answerId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//从elastSearch删掉该数据
	err2 := elastSearch.DelAnswerFromEs(answerId)
	if err2 != nil {
		log.Errorf(err2.Error())
	}

	//从经验弹中删除该记录 不需要校验经验弹权限  TODO (删除更改experience_share 的status状态  或者删除记录)
	experience.DelAnswerFromExpByAnswerId(uin, answerId)

	if !isMyself {
		answer, _ := common.GetV2Answer(answerId)
		data, err1 := json.Marshal(&answer)
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}
		dataStr := string(data)
		if uin != qidOwner {
			v2push.SendBeDeletePush(uin, ownerUid, reason, 2, dataStr)
		}
	}

	code = 0

	log.Debugf("end DelAnswer uin = %d qid = %d answerId = %d code = %d", uin, qid, answerId, code)
	return
}

func getDelAnswerPermitOperators(answerId, qid int) (operators []int64, owner int64, qidOwner int64, err error) {

	if answerId == 0 || qid == 0 {
		log.Errorf("qid or answerId is zero")
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

	//本版块的墙主有权限删除本版块的回答
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

	//回答者本人有权删除自己的回答  问题所有者有权利删除回答
	//var owner int64
	sql = fmt.Sprintf(`select v2questions.ownerUid,v2answers.ownerUid from v2questions,v2answers 
where  v2answers.answerId = %d and v2answers.qid = %d and v2questions.qid = %d`, answerId, qid, qid)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&qidOwner, &owner)
	}

	if qidOwner != 0 {
		operators = append(operators, qidOwner)
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
