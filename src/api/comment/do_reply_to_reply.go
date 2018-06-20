package comment

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type ReplyToReplyReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId     int    `schema:"answerId"`
	CommentId    int    `schema:"commentId"`
	ReplyId      int    `schema:"replyId"`
	ReplyContent string `schema:"replyContent"`
}

type ReplyToReplyRsp struct {
	ReplyId int `json:"replyId"`
}

func doReplyToReply(req *ReplyToReplyReq, r *http.Request) (rsp *ReplyToReplyRsp, err error) {

	log.Debugf("uin %d, ReplyToReplyReq %+v", req.Uin, req)

	replyId, err := ReplyToReply(req.Uin, req.AnswerId, req.CommentId, req.ReplyId, req.ReplyContent)

	if err != nil {
		log.Errorf("uin %d, PostAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ReplyToReplyRsp{int(replyId)}

	log.Debugf("uin %d, ReplyToCommentRsp succ, %+v", req.Uin, rsp)

	return
}

func ReplyToReply(uin int64, answerId, commentId, replyId int, replyContent string) (repId int64, err error) {

	if answerId <= 0 || commentId <= 0 || replyId <= 0 || len(replyContent) == 0 {
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

	sql := fmt.Sprintf(`select fromUid from v2replys where replyId = %d`, replyId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var toUid int64
	for rows.Next() {
		rows.Scan(&toUid)
	}

	stmt, err := inst.Prepare(`insert into v2replys(replyId, commentId, replyContent, fromUid, toUid, replyStatus, replyTs) 
		values(?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	res, err := stmt.Exec(0, commentId, replyContent, uin, toUid, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	repId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
