package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"sort"
	"svr/st"
)

type GetAnswersReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid      int `schema:"qid"`
	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetAnswersRsp struct {
	Answers  []*st.AnswersInfo `json:"answers"`
	TotalCnt int               `json:"totalCnt"`
}

// 自定义排序 按照赞的多少
type answerSort []*st.AnswersInfo

func (I answerSort) Len() int {
	return len(I)
}

func (I answerSort) Less(i, j int) bool {
	return I[i].LikeCnt > I[j].LikeCnt
}

func (I answerSort) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

func doGetAnswers(req *GetAnswersReq, r *http.Request) (rsp *GetAnswersRsp, err error) {

	//log.Debugf("uin %d, GetAnswersReq %+v", req.Uin, req)

	answers, totalCnt, err := GetAnswers(req.Uin, req.Qid, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetAnswers error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetAnswersRsp{answers, totalCnt}

	//log.Debugf("uin %d, GetQuestionsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetAnswers(uin int64, qid, pageNum, pageSize int) (answers []*st.AnswersInfo, totalCnt int, err error) {

	//log.Debugf("start GetAnswers uin:%d", uin)

	if qid <= 0 || pageNum < 0 || pageSize < 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}
	answers = make([]*st.AnswersInfo, 0)

	expAnswer := make([]*st.AnswersInfo, 0)
	otherAnswer := make([]*st.AnswersInfo, 0)

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := s + pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(answerId) as cnt from v2answers where qid = %d and  answerStatus = 0`, qid)
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

	sql = fmt.Sprintf(`select qid, ownerUid, answerId, answerContent, answerImgUrls, answerTs  from v2answers where answerStatus = 0 and qid = %d order by answerTs desc`, qid)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.AnswersInfo
		var uid int64

		rows.Scan(
			&info.Qid,
			&uid,
			&info.AnswerId,
			&info.AnswerContent,
			&info.AnswerImgUrls,
			&info.AnswerTs)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		//TODO 增加最新回复的两条评论  2018-07-19改为最早两条
		latestComments, err2 := GetLatesCommentByAnswerId(info.AnswerId)
		info.LatestComment = latestComments
		if err2 == nil {
			info.LatestComment = latestComments
		}

		//查找该问题的labelName
		expLabels, err3 := GetLabelInfoByAnswerId(info.AnswerId)
		if err3 == nil {
			info.ExpLabel = expLabels
		}

		commentCnt, err1 := getCommentCnt(info.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		info.CommentCnt = commentCnt

		//点赞数

		likeCnt, err1 := getAnswerLikeCnt(info.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		info.LikeCnt = likeCnt

		isILike, err1 := checkIsILikeAnswer(uin, info.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		info.IsILike = isILike

		//分成两个slice
		if len(expLabels) > 0 {
			expAnswer = append(expAnswer, &info)
		} else {
			otherAnswer = append(otherAnswer, &info)
		}

		//answers = append(answers, &info)
	}

	sort.Sort(answerSort(expAnswer))
	sort.Sort(answerSort(otherAnswer))

	for _, tmp := range expAnswer {
		var answerTmp = tmp
		answers = append(answers, answerTmp)
	}

	for _, tmp := range otherAnswer {
		var answerTmp = tmp
		answers = append(answers, answerTmp)
	}
	//copy(answers,expAnswer)
	//copy(answers[len(expAnswer):],otherAnswer)

	//retAnsers, err := sortQuestionAnswer(answers)
	//if err != nil {
	//	log.Error(err.Error())
	//	return
	//}
	//
	//// 超过最大长度  等于slice长度
	//totalCnt = len(retAnsers)
	//if e > totalCnt {
	//	e = totalCnt
	//}
	//
	////s - e
	//retAnsers = retAnsers[s:e]
	//
	//log.Debugf("end GetAnswers uin:%d totalCnt:%d, len:%d ,s:%d,e:%d", uin, totalCnt, len(retAnsers), s, e)

	totalCnt = len(answers)
	if e > totalCnt {
		e = totalCnt
	}
	answers = answers[s:e]
	log.Debugf("end GetAnswers uin:%d totalCnt:%d, len:%d ,s:%d,e:%d", uin, totalCnt, len(answers), s, e)
	return
}

func sortQuestionAnswer(answers []*st.AnswersInfo) (sortedAnswers []*st.AnswersInfo, err error) {

	//log.Debugf("start sortQuestionAnswer")

	likeCntAnswerMap := make(map[int][]*st.AnswersInfo)

	for _, answer := range answers {
		if _, ok := likeCntAnswerMap[answer.LikeCnt]; !ok {
			likeCntAnswerMap[answer.LikeCnt] = make([]*st.AnswersInfo, 0)
		}
		likeCntAnswerMap[answer.LikeCnt] = append(likeCntAnswerMap[answer.LikeCnt], answer)
	}

	likeCntSet := make([]int, 0)
	for key := range likeCntAnswerMap {
		likeCntSet = append(likeCntSet, key)
	}

	sort.Ints(likeCntSet[:])

	for i := len(likeCntSet) - 1; i >= 0; i-- {
		for _, answer := range likeCntAnswerMap[likeCntSet[i]] {
			sortedAnswers = append(sortedAnswers, answer)
		}
	}

	//log.Debugf("end sortQuestionAnswer sortedAnswers:%+v", sortedAnswers)
	return
}

func getCommentCnt(answerId int) (cnt int, err error) {
	//log.Debugf("start getCommentCnt answerId:%d", answerId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//2018-07-05 评论数 = 评论数 + 回复数
	sql := fmt.Sprintf(`select commentId  from v2comments where answerId = %d and commentStatus = 0`, answerId)
	rows, err := inst.Query(sql)
	defer rows.Close()

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	commentIdArr := make([]int, 0)
	commentCnt := 0
	for rows.Next() {
		var commentId int
		rows.Scan(&commentId)
		commentCnt++
		commentIdArr = append(commentIdArr, commentId)
	}
	commentStr := ""
	for i, commentId := range commentIdArr {
		if i != len(commentIdArr)-1 {
			commentStr += fmt.Sprintf(`%d,`, commentId)
		} else {
			commentStr += fmt.Sprintf(`%d`, commentId)
		}
	}

	//找到每一条评论的回复数
	sql = fmt.Sprintf(`select count(*) as cnt from v2replys where commentId in ('%s') and replyStatus = 0 `, commentStr)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	replyCnt := 0
	for rows.Next() {
		rows.Scan(&replyCnt)
	}
	cnt = replyCnt + commentCnt

	/*
		// 评论数
		sql := fmt.Sprintf(`select count(commentId) as cnt from v2comments where answerId = %d and commentStatus = 0`, answerId)
	*/

	//log.Debugf("end getCommentCnt answerId:%d commentCnt:%d,replyCnt:%d,cnt:%d", answerId, commentCnt, replyCnt, cnt)
	return
}

func getAnswerLikeCnt(answerId int) (cnt int, err error) {
	//log.Debugf("start getAnswerLikeCnt answerId:%d", answerId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	// 点赞数
	sql := fmt.Sprintf(`select count(id) as cnt from v2likes where type = 1 and likeId = %d and likeStatus != 2`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&cnt)
	}

	//log.Debugf("end getAnswerLikeCnt answerId:%d cnt:%d", answerId, cnt)
	return
}

func checkIsILikeAnswer(uin int64, answerId int) (ret bool, err error) {
	//log.Debugf("start checkIsILikeAnswer uin:%d answerId:%d", uin, answerId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select id from v2likes where type = 1 and likeId = %d and ownerUid = %d and likeStatus != 2`, answerId, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		rows.Scan(&id)
		ret = true
	}

	//log.Debugf("end checkIsILikeAnswer uin:%d answerId:%d ret:%t", uin, answerId, ret)
	return
}

func GetLatesCommentByAnswerId(answerId int) (comments []*st.CommentInfo, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select * from v2comments where answerId = %d  limit 2`, answerId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.CommentInfo
		var uid int64
		var ts int
		info.Replys = make([]st.ReplyInfo, 0)

		rows.Scan(
			&info.CommentId,
			&info.AnswerId,
			&info.CommentContent,
			&uid,
			&ts,
			&info.CommentTs)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		comments = append(comments, &info)
	}

	return
}

func GetLabelInfoByAnswerId(answerId int) (expLabels []*st.ExpLabel, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select experience_label.labelId, experience_label.labelName from experience_share,experience_label 
where experience_share.labelId = experience_label.labelId and experience_share.answerId = %d`, answerId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var expLabel st.ExpLabel
		rows.Scan(&expLabel.LabelId, &expLabel.LabelName)
		expLabels = append(expLabels, &expLabel)
	}

	return
}
