package comment

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type ReplyToCommentReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId     int    `schema:"answerId"`
	CommentId    int    `schema:"commentId"`
	ReplyContent string `schema:"replyContent"`
}

type ReplyToCommentRsp struct {
	ReplyId int `json:"replyId"`
}

func doReplyToComment(req *ReplyToCommentReq, r *http.Request) (rsp *ReplyToCommentRsp, err error) {

	log.Debugf("uin %d, ReplyToCommentReq %+v", req.Uin, req)

	replyId, err := ReplyToComment(req.Uin, req.AnswerId, req.CommentId, req.ReplyContent)

	if err != nil {
		log.Errorf("uin %d, PostAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ReplyToCommentRsp{int(replyId)}

	log.Debugf("uin %d, ReplyToCommentRsp succ, %+v", req.Uin, rsp)

	return
}

func ReplyToComment(uin int64, answerId, commentId int, replyContent string) (replyId int64, err error) {

	if answerId <= 0 || commentId <= 0 || len(replyContent) == 0 {
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

	sql := fmt.Sprintf(`select ownerUid from v2comments where commentId = %d`, commentId)
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

	replyId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
