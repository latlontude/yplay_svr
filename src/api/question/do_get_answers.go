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

	sql := fmt.Sprintf(`select count(answerId) as cnt from v2answers where qid = %d`, qid)
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

		answers = append(answers, &info)
	}

	retAnsers, err := sortQuestionAnswer(answers)
	if err != nil {
		log.Error(err.Error())
		return
	}

	answers = make([]*st.AnswersInfo, 0)
	for i := 0; i < len(retAnsers); i++ {
		if i >= s {
			answers = append(answers, retAnsers[i])
		}

		if i > e-1 {
			break
		}
	}

	//log.Debugf("end GetAnswers uin:%d totalCnt:%d", uin, totalCnt)
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
	for key, _ := range likeCntAnswerMap {
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

	// 评论数
	sql := fmt.Sprintf(`select count(commentId) as cnt from v2comments where answerId = %d`, answerId)
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

	//log.Debugf("end getCommentCnt answerId:%d cnt:%d", answerId, cnt)
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
	sql := fmt.Sprintf(`select count(id) as cnt from v2likes where type = 1 and likeId = %d`, answerId)
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

	sql := fmt.Sprintf(`select id from v2likes where type = 1 and likeId = %d and ownerUid = %d`, answerId, uin)
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
