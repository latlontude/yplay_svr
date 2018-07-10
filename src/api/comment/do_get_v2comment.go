package comment

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func GetV2Comment(commentId int) (comment st.CommentInfo, ownerUid int64, err error) {
	if commentId == 0 {
		log.Errorf("commentId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select commentId ,answerId,commentContent,ownerUid,commentTs from v2comments where commentId = %d`,
		commentId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var uid int64

	for rows.Next() {
		rows.Scan(&comment.CommentId, &comment.AnswerId, &comment.CommentContent, &uid, &comment.CommentTs)
	}

	ownerUid = uid
	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		comment.OwnerInfo = ui
	}

	return
}


//replyId replyInfo
//ownerUid 该回复归属者

func GetV2Reply(replyId int) (reply st.ReplyInfo, ownerUid int64, err error) {
	if replyId == 0 {
		log.Errorf("replyId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select replyId,replyContent,fromUid,toUid,replyTs  from v2replys where replyId = %d`, replyId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	var fromUid int64
	var toUid int64

	for rows.Next() {
		rows.Scan(&reply.ReplyId, &reply.ReplyContent, &fromUid, &toUid, &reply.ReplyTs)
	}

	//被点赞对象
	ownerUid = toUid

	if fromUid > 0 {
		ui, err1 := st.GetUserProfileInfo(fromUid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		reply.ReplyFromUserInfo = ui
	}

	if toUid > 0 {
		ui, err1 := st.GetUserProfileInfo(toUid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		reply.ReplyToUserInfo = ui
	}

	return
}

