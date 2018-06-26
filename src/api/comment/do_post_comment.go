package comment

import (
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"strings"
	"svr/st"
	"time"
)

type PostCommentReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId    int    `schema:"answerId"`
	CommentText string `schema:"commentText"`
}

type PostCommentRsp struct {
	CommentId int `json:"commentId"`
}

func doPostComment(req *PostCommentReq, r *http.Request) (rsp *PostCommentRsp, err error) {

	log.Debugf("uin %d, PostCommentReq %+v", req.Uin, req)

	commentId, err := PostComment(req.Uin, req.AnswerId, strings.Trim(req.CommentText, " \n\t"))

	if err != nil {
		log.Errorf("uin %d, PostComment error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PostCommentRsp{int(commentId)}

	log.Debugf("uin %d, PostQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func PostComment(uin int64, answerId int, commentText string) (commentId int64, err error) {

	log.Debugf("start PostComment uin:%d", uin)

	if answerId == 0 || len(commentText) == 0 {
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

	stmt, err := inst.Prepare(`insert into v2comments(commentId, answerId, commentContent, ownerUid, commentStatus, commentTs) 
		values(?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	res, err := stmt.Exec(0, answerId, commentText, uin, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	commentId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	var newComment st.CommentInfo
	newComment.AnswerId = answerId
	newComment.CommentId = int(commentId)
	newComment.CommentContent = commentText
	newComment.CommentTs = int(ts)
	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}
	newComment.OwnerInfo = ui

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16
	go v2push.SendBeCommentPush(answerId, newComment)

	log.Debugf("end PostComment uin:%d commentId:%d", uin, commentId)
	return
}
