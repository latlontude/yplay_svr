package comment

import (
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelCommentReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId  int `schema:"answerId"`
	CommentId int `schema:"commentId"`
	Reason    string `schema:"reason"`
}

type DelCommentRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelComment(req *DelCommentReq, r *http.Request) (rsp *DelCommentRsp, err error) {

	log.Debugf("uin %d, DelCommentReq %+v", req.Uin, req)

	code, err := DelComment(req.Uin, req.AnswerId, req.CommentId ,req.Reason)

	if err != nil {
		log.Errorf("uin %d, DelAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelCommentRsp{code}

	log.Debugf("uin %d, DelCommentRsp succ, %+v", req.Uin, rsp)

	return
}

func DelComment(uin int64, answerId, commentId int, reason string) (code int, err error) {
	log.Debugf("start DelComment uin = %d answerId = %d commentId = %d", uin, answerId, commentId)

	code = -1

	if answerId <= 0 || commentId <= 0 {
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

	uids, ownerUid,err := getDelCommentPermitOperators(answerId, commentId)
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
		log.Errorf("uin:%d has no permissions to delete comment:%d", uin, commentId)
		return
	}

	sql := fmt.Sprintf(`delete from v2comments where answerId = %d and commentId = %d`, answerId, commentId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//不是我自己删的 发推送
	if !isMyself {
		SendBeDeletePush(uin, ownerUid ,reason, 3)
	}

	code = 0

	log.Debugf("end DelComment uin = %d  code = %d", uin, code)
	return
}

func getDelCommentPermitOperators(answerId, commentId int) (operators []int64, owner int64, err error) {

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

	//本版块的墙主有权限删除本版块的评论
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

	//评论者本人有权删除自己的评论
	//var owner int64
	sql = fmt.Sprintf(`select ownerUid from v2comments where commentId = %d and answerId = %d`, commentId, answerId)
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
