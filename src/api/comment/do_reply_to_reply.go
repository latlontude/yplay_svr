package comment

import (
	"api/v2push"
	_ "api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type ReplyToReplyReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid          int    `schema:"qid"`
	AnswerId     int    `schema:"answerId"`
	CommentId    int    `schema:"commentId"`
	ReplyId      int    `schema:"replyId"`
	ReplyContent string `schema:"replyContent"`
	Ext          string `schema:"ext"`
	ToUin        int64  `schema:"toUin"`
}

type ReplyToReplyRsp struct {
	ReplyId int `json:"replyId"`
}

func doReplyToReply(req *ReplyToReplyReq, r *http.Request) (rsp *ReplyToReplyRsp, err error) {

	log.Debugf("uin %d, ReplyToReplyReq %+v", req.Uin, req)

	replyId, err := V2ReplayToComment(req.Uin, req.ToUin, req.Qid, req.AnswerId, req.CommentId, req.ReplyContent, req.Ext)

	if err != nil {
		log.Errorf("uin %d, PostAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ReplyToReplyRsp{int(replyId)}

	log.Debugf("uin %d, ReplyToCommentRsp succ, %+v", req.Uin, rsp)

	return
}

func V2ReplayToComment(uin int64, toUin int64, qid int, answerId int, commentId int, replyContent string, ext string) (newCommentId int64, err error) {

	if (answerId <= 0 || len(replyContent) == 0) && len(ext) == 0 {
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

	if toUin == 0 {
		sql := fmt.Sprintf(`select ownerUid from v2comments where commentId = %d`, commentId)
		rows, err2 := inst.Query(sql)
		if err2 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err2.Error())
			log.Error(err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&toUin)
		}
	}

	//插入到新的表
	stmt, err := inst.Prepare(`insert into v2comments(commentId, answerId, commentContent, ownerUid, toUid, isAnonymous,commentStatus, commentTs,ext)
	values(?, ?, ?, ?, ?, ?, ?, ? ,?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	res, err := stmt.Exec(0, answerId, replyContent, uin, toUin, 0, status, ts, ext)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	newCommentId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	var newComment st.CommentInfo
	newComment.AnswerId = answerId
	newComment.CommentId = int(newCommentId)
	newComment.CommentContent = replyContent
	newComment.CommentTs = int(ts)
	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if toUin > 0 {
		toUi, _ := st.GetUserProfileInfo(toUin)
		newComment.ToOwnerInfo = toUi
	}
	newComment.OwnerInfo = ui

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16

	if len(ext) > 0 && ext != "null" {
		go v2push.SendAtPush(uin, 3, qid, newComment, ext)
	} else {
		go v2push.SendV2BeCommentPush(uin, toUin, qid, answerId, newComment)
	}
	log.Debugf("end PostComment uin:%d commentId:%d", uin, commentId)
	return

}
func ReplyToReply(uin int64, qid, answerId, commentId, replyId int, replyContent string, ext string) (repId int64, err error) {

	if (answerId <= 0 || commentId <= 0 || replyId <= 0 || len(replyContent) == 0) && len(ext) == 0 {
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

	//新生成的replyId
	repId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	var newReply st.ReplyInfo
	newReply.ReplyId = int(repId)
	newReply.ReplyContent = replyContent
	newReply.ReplyTs = ts

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}
	newReply.ReplyFromUserInfo = ui

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16

	if len(ext) > 0 && ext != "null" {
		go v2push.SendAtPush(uin, 4, qid, newReply, ext)

	} else {
		go v2push.SendReplyBeReplyPush(uin, qid, answerId, commentId, int(repId), newReply)
	}

	return
}
