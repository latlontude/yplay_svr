package comment

import (
	"api/common"
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
)

type DelReplyReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId  int    `schema:"answerId"`
	CommentId int    `schema:"commentId"`
	ReplyId   int    `schema:"replyId"`
	Reason    string `schema:"reason"`
}

type DelReplyRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelReply(req *DelReplyReq, r *http.Request) (rsp *DelReplyRsp, err error) {

	log.Debugf("uin %d, DelReplyReq %+v", req.Uin, req)

	code, err := DelReply(req.Uin, req.AnswerId, req.CommentId, req.ReplyId, req.Reason)

	if err != nil {
		log.Errorf("uin %d, DelReply error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelReplyRsp{code}

	log.Debugf("uin %d, DelReplyRsp succ, %+v", req.Uin, rsp)

	return
}

func DelReply(uin int64, answerId, commentId, replyId int, reason string) (code int, err error) {
	log.Debugf("start DelReply uin = %d", uin)

	code = -1

	if answerId <= 0 || commentId <= 0 || replyId <= 0 {
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
	uids, ownerUid, err := getDelReplyPermitOperators(answerId, commentId, replyId)
	if err != nil {
		log.Error(err)
		return
	}
	isMyself := false
	permit := false
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
		log.Errorf("uin:%d has no permissions to delete replyId:%d", uin, replyId)
		return
	}

	sql := fmt.Sprintf(`update v2replys set replyStatus = 1 where replyId = %d and commentId = %d`, replyId, commentId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//不是我自己删的 发推送
	if !isMyself {
		reply, _, _ := common.GetV2Reply(replyId)
		data, err1 := json.Marshal(&reply)
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}
		dataStr := string(data)
		v2push.SendBeDeletePush(uin, ownerUid, reason, 4, dataStr)
	}

	code = 0

	log.Debugf("end DelReplay uin = %d code = %d", uin, code)
	return
}

func getDelReplyPermitOperators(answerId, commentId, replyId int) (operators []int64, owner int64, err error) {

	if answerId == 0 || commentId == 0 {
		log.Errorf("commentId or answerId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	qid := 0
	sql := fmt.Sprintf(`select qid from v2answers where answerId = %d`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	for rows.Next() {
		rows.Scan(&qid)
	}

	boardId := 0
	sql = fmt.Sprintf(`select boardId from v2questions where qid = %d`, qid)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	for rows.Next() {
		rows.Scan(&boardId)
	}

	//本版块的墙主有权限删除本版块的回应
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

	//回应者本人有权删除自己的回应
	//var owner int64
	sql = fmt.Sprintf(`select fromUid from v2replys where replyId = %d and commentId = %d`, replyId, commentId)
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
