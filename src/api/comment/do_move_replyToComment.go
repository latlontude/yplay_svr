package comment

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type MoveReq struct {
}

type MoveRsp struct {
	Code int `json:"code"`
}

func doMoveReplyToComment(req *MoveReq, r *http.Request) (rsp *MoveRsp, err error) {

	log.Debugf("uin %d, MoveReq %+v", req)
	err = MoveReplyToComment()
	if err != nil {
		log.Errorf(" MoveReq error, %s", err.Error())
		return
	}

	code := 0
	rsp = &MoveRsp{code}

	return
}

// 新的评论 2018-09-17 只有 问题 回答 评论 三级

func MoveReplyToComment() (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select 
		v2comments.answerId,
		v2replys.replyContent,
		v2replys.fromUid,
		v2replys.toUid ,
		v2comments.isAnonymous,
		v2replys.replyStatus,
		v2replys.replyTs,
		v2replys.ext
		from v2replys,v2comments  where v2replys.commentId = v2comments.commentId`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	commentIdList := make([]int64, 0)
	for rows.Next() {
		var uin, toUin int64
		var answerId int
		var commentText string
		var isAnonymous bool
		var status int
		var ts int64
		var ext string
		rows.Scan(&answerId, &commentText, &uin, &toUin, &isAnonymous, &status, &ts, &ext)

		stmt, err2 := inst.Prepare(`insert into v2comments(commentId, answerId, commentContent,
ownerUid, toUid, isAnonymous,commentStatus, commentTs,ext) values(?, ?, ?, ?, ?, ?, ?, ? ,?)`)

		if err2 != nil {
			err = rest.NewAPIError(constant.E_DB_PREPARE, err2.Error())
			log.Error(err.Error())
			return
		}

		res, err3 := stmt.Exec(0, answerId, commentText, uin, toUin, isAnonymous, status, ts, 0)
		if err3 != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err3.Error())
			log.Error(err.Error())
			return
		}

		commentId, err4 := res.LastInsertId()
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err4.Error())
			log.Error(err.Error())
			return
		}
		commentIdList = append(commentIdList, commentId)

	}
	log.Debugf("commentList:%+v", commentIdList)
	return
}
