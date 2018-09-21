package answer

import (
	"api/common"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetCommentsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId int `schema:"answerId"`
	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetCommentsRsp struct {
	Comments []*st.CommentInfo `json:"comments"`
	TotalCnt int               `json:"totalCnt"`
}

func doGetComments(req *GetCommentsReq, r *http.Request) (rsp *GetCommentsRsp, err error) {

	log.Debugf("uin %d, GetCommentsReq %+v", req.Uin, req)

	comments, totalCnt, err := GetComments(req.Uin, req.AnswerId, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetQuestions error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetCommentsRsp{comments, totalCnt}

	log.Debugf("uin %d, GetCommentsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetComments(uin int64, answerId, pageNum, pageSize int) (comments []*st.CommentInfo, totalCnt int, err error) {

	//log.Debugf("start GetComments uin:%d", uin)

	if answerId <= 0 || pageNum < 0 || pageSize < 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}
	comments = make([]*st.CommentInfo, 0)

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(commentId) as cnt from  v2comments where answerId = %d and commentStatus = 0`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&totalCnt)
	}

	if totalCnt == 0 {
		return
	}

	sql = fmt.Sprintf(`select commentId, answerId, commentContent, ownerUid, toUid,isAnonymous,commentTs ,ext from v2comments where commentStatus = 0 
and answerId = %d order by commentTs  limit %d, %d`, answerId, s, e)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.CommentInfo
		var uid, toUid int64

		info.Replys = make([]st.ReplyInfo, 0)

		rows.Scan(
			&info.CommentId,
			&info.AnswerId,
			&info.CommentContent,
			&uid,
			&toUid,
			&info.IsAnonymous,
			&info.CommentTs,
			&info.Ext)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		if toUid > 0 {
			toUi, err1 := st.GetUserProfileInfo(toUid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.ToOwnerInfo = toUi
		}

		//Replys, err1 := getReplyArray(uin, info.CommentId)
		//if err1 != nil {
		//	log.Error(err1.Error())
		//	continue
		//}
		//
		//info.Replys = Replys

		commentLikeCnt, err1 := common.GetLikeCntByType(info.CommentId, 2)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		info.LikeCnt = commentLikeCnt

		isILikeComment, err1 := common.CheckIsILike(uin, info.CommentId, 2)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		info.IsILike = isILikeComment

		comments = append(comments, &info)
	}

	//log.Debugf("end GetComments uin:%d totalCnt:%d", uin, totalCnt)
	return
}

func getReplyArray(uin int64, commentId int) (Replys []st.ReplyInfo, err error) {
	//log.Debugf("start getReplyArray uin:%d commentId:%d", uin, commentId)

	Replys = make([]st.ReplyInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select replyId, replyContent, fromUid, toUid, replyTs ,ext from v2replys where commentId = %d and replyStatus = 0`, commentId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var fromUid int64
		var toUid int64
		var replyInfo st.ReplyInfo

		rows.Scan(
			&replyInfo.ReplyId,
			&replyInfo.ReplyContent,
			&fromUid,
			&toUid,
			&replyInfo.ReplyTs,
			&replyInfo.Ext)

		if fromUid > 0 {
			ui, err1 := st.GetUserProfileInfo(fromUid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			replyInfo.ReplyFromUserInfo = ui
		}

		if toUid > 0 {
			ui, err1 := st.GetUserProfileInfo(toUid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			replyInfo.ReplyToUserInfo = ui
		}

		replyLikeCnt, err1 := common.GetLikeCntByType(replyInfo.ReplyId, 3)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		replyInfo.LikeCnt = replyLikeCnt

		isILikeReply, err1 := common.CheckIsILike(uin, replyInfo.ReplyId, 3)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		replyInfo.IsILike = isILikeReply

		Replys = append(Replys, replyInfo)
	}

	//log.Debugf("end getReplyArray uin:%d commentId:%d replayInfos:%+v", uin, commentId, Replys)
	return
}
