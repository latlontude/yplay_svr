package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetQuestionsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int `schema:"boardId"`
	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetQuestionsRsp struct {
	V2Questions []*st.V2QuestionInfo `json:"questions"`
	TotalCnt    int                  `json:"totalCnt"`
}

func doGetQuestions(req *GetQuestionsReq, r *http.Request) (rsp *GetQuestionsRsp, err error) {

	log.Debugf("uin %d, GetQuestionsReq %+v", req.Uin, req)

	questions, totalCnt, err := GetQuestions(req.Uin, req.BoardId, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetQuestions error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetQuestionsRsp{questions, totalCnt}

	log.Debugf("uin %d, GetQuestionsRsp succ", req.Uin)

	return
}

func GetQuestions(uin int64, boardId, pageNum, pageSize int) (questions []*st.V2QuestionInfo, totalCnt int, err error) {

	log.Debugf("start GetQuestions uin:%d", uin)

	if boardId <= 0 || pageNum < 0 || pageSize < 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}
	questions = make([]*st.V2QuestionInfo, 0)

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

	sql := fmt.Sprintf(`select count(qid) as cnt from  v2questions where boardId = %d and qStatus = 0`, boardId)
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

	sql = fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls, isAnonymous, createTs, modTs from v2questions where qStatus = 0 and (qContent != "" or qImgUrls != "") and boardId = %d order by createTs desc limit %d, %d`, boardId, s, e)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.V2QuestionInfo
		var uid int64

		rows.Scan(
			&info.Qid,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.IsAnonymous,
			&info.CreateTs,
			&info.ModTs)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		answerCnt, err := getAnswerCnt(info.Qid)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		info.AnswerCnt = answerCnt

		bestAnswer, _ := getBestAnswer(uin, info.Qid)
		info.BestAnswer = bestAnswer

		responders, _ := getQidNewResponders(info.Qid)
		info.NewResponders = responders

		questions = append(questions, &info)
	}

	log.Debugf("end GetQuestions uin:%d totalCnt:%d", uin, totalCnt)
	return
}

func getAnswerCnt(qid int) (cnt int, err error) {
	log.Debugf("start getAnswerCnt qid:%d", qid)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(answerId) as cnt from v2answers where qid = %d and answerStatus = 0`, qid)
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

	log.Debugf("end getAnswerCnt qid:%d totalCnt:%d", qid, cnt)
	return
}

func getBestAnswer(uin int64, qid int) (answer *st.AnswersInfo, err error) {
	log.Debugf("start getBestAnswer qid:%d", qid)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(answerId) as cnt from v2answers where qid = %d and answerStatus = 0`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	totalCnt := 0
	for rows.Next() {
		rows.Scan(&totalCnt)
	}

	if totalCnt == 0 {
		return
	}

	answers := make([]*st.AnswersInfo, 0)

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

	sortAnswers, err := sortQuestionAnswer(answers)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if len(sortAnswers) > 0 {
		if sortAnswers[0].LikeCnt != 0 {
			answer = sortAnswers[0]
		}
	}

	log.Debugf("end getBestAnswer answer:%+v", answer)
	return
}

func getQidNewResponders(qid int) (responders []*st.UserProfileInfo, err error) {

	if qid == 0 {
		log.Errorf("qid is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//查找本道题目最新回答的两个人
	sql := fmt.Sprintf(`select ownerUid from v2answers where qid = %d and answerStatus = 0 order by answerTs desc limit 2`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			responders = append(responders, ui)
		}
	}

	return
}
