package comment

import (
	"api/v2push"
	_ "api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"strings"
	"svr/st"
	"time"
)

type ReplyToCommentReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid          int    `schema:"qid"`
	AnswerId     int    `schema:"answerId"`
	CommentId    int    `schema:"commentId"`
	ReplyContent string `schema:"replyContent"`
	Ext          string `schema:"ext"`
}

type ReplyToCommentRsp struct {
	ReplyId int `json:"replyId"`
}

func doReplyToComment(req *ReplyToCommentReq, r *http.Request) (rsp *ReplyToCommentRsp, err error) {

	log.Debugf("uin %d, ReplyToCommentReq %+v", req.Uin, req)

	replyId, err := ReplyToComment(req.Uin, req.Qid, req.AnswerId, req.CommentId, strings.Trim(req.ReplyContent, " \n\t"), req.Ext)

	if err != nil {
		log.Errorf("uin %d, PostAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ReplyToCommentRsp{int(replyId)}

	log.Debugf("uin %d, ReplyToCommentRsp succ, %+v", req.Uin, rsp)

	return
}

func ReplyToComment(uin int64, qid, answerId, commentId int, replyContent string, ext string) (replyId int64, err error) {

	if (answerId <= 0 || commentId <= 0 || len(replyContent) == 0) && len(ext) == 0 {
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

	stmt, err := inst.Prepare(`insert into v2replys(replyId, commentId, replyContent, fromUid, toUid, replyStatus, replyTs,ext) 
		values(?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	res, err := stmt.Exec(0, commentId, replyContent, uin, toUid, status, ts, ext)
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

	var newReply st.ReplyInfo
	newReply.ReplyId = int(replyId)
	newReply.ReplyContent = replyContent
	newReply.ReplyTs = ts

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}
	newReply.ReplyFromUserInfo = ui

	//给评论者发送push，告诉ta，ta的回答收到了新评论 dataType:16

	if len(ext) > 0 {
		go v2push.SendAtPush(uin, 4, qid, newReply, ext)
	} else {
		go v2push.SendCommentBeReplyPush(uin, qid, answerId, commentId, newReply)
	}

	return
}
