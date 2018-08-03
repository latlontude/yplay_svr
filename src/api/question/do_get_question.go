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

type GetQuestionsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int `schema:"boardId"`
	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
	Qid      int `schema:"lastQid"`
}

type GetQuestionsRsp struct {
	V2Questions []*st.V2QuestionInfo `json:"questions"`
	TotalCnt    int                  `json:"totalCnt"`
}

func doGetQuestions(req *GetQuestionsReq, r *http.Request) (rsp *GetQuestionsRsp, err error) {

	log.Debugf("uin %d, GetQuestionsReq %+v", req.Uin, req)

	questions, totalCnt, err := GetQuestions(req.Uin, req.Qid, req.BoardId, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetQuestions error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetQuestionsRsp{questions, totalCnt}

	log.Debugf("uin %d, GetQuestionsRsp succ  , rsp:%v", req.Uin, rsp)

	return
}

func GetQuestions(uin int64, qid, boardId, pageNum, pageSize int) (questions []*st.V2QuestionInfo, totalCnt int, err error) {

	//log.Debugf("start GetQuestions uin:%d", uin)

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

	//第一次拉去列表 没有qid
	if qid == 0 {
		sql = fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls, isAnonymous, createTs, modTs ,ext from v2questions 
		where qStatus = 0 and (qContent != "" or qImgUrls != "") and boardId = %d order by createTs desc limit %d, %d`, boardId, s, e)
	} else {
		//后面拉去问题列表防止插入 重复数据 客户端传qid,从小于qid的地方去pageSize
		sql = fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls, isAnonymous, createTs, modTs ,ext from v2questions 
		where qStatus = 0 and (qContent != "" or qImgUrls != "") and boardId = %d  and qid < %d
		order by qid desc limit %d, %d`, boardId, qid, s, e)
	}

	log.Errorf("SQL:%s", sql)
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
			&info.ModTs,
			&info.Ext)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		//answerCnt, err := GetAnswerCnt(info.Qid)

		answerCnt, err := GetDiscussCnt(info.Qid)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		info.AnswerCnt = answerCnt

		bestAnswer, _ := GetBestAnswer(uin, info.Qid)
		info.BestAnswer = bestAnswer
		if bestAnswer != nil {
			responders, _ := GetQidNewResponders(info.Qid, bestAnswer)
			info.NewResponders = responders
		}

		questions = append(questions, &info)
	}

	//log.Debugf("end GetQuestions uin:%d totalCnt:%d", uin, totalCnt)
	return
}

func GetAnswerCnt(qid int) (cnt int, err error) {
	//log.Debugf("start getAnswerCnt qid:%d", qid)

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

	//log.Debugf("end getAnswerCnt qid:%d totalCnt:%d", qid, cnt)
	return
}

func GetBestAnswer(uin int64, qid int) (answer *st.AnswersInfo, err error) {
	//log.Debugf("start getBestAnswer qid:%d", qid)

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
	expAnswer := make([]*st.AnswersInfo, 0)   //带经验弹的回答
	otherAnswer := make([]*st.AnswersInfo, 0) //不带经验弹的回答
	sql = fmt.Sprintf(`select qid, ownerUid, answerId, answerContent, answerImgUrls, answerTs  ,ext from v2answers where answerStatus = 0 and qid = %d order by answerTs desc`, qid)

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
			&info.AnswerTs,
			&info.Ext)

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

		//查找该问题的labelName
		expLabels, err3 := GetLabelInfoByAnswerId(info.AnswerId)
		if err3 == nil {
			info.ExpLabel = expLabels
		}
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

	if len(answers) > 0 {
		answer = answers[0]
	}

	//sortAnswers, err := sortQuestionAnswer(answers)
	//if err != nil {
	//	log.Error(err.Error())
	//	return
	//}
	//
	//if len(sortAnswers) > 0 {
	//	if sortAnswers[0].LikeCnt != 0 {
	//		answer = sortAnswers[0]
	//	}
	//}
	//	log.Debugf("end getBestAnswer answer:%+v", answer)
	return
}

func GetQidNewResponders(qid int, bestAnswer *st.AnswersInfo) (responders []*st.UserProfileInfo, err error) {

	if qid == 0 {
		log.Errorf("qid is zero")
		return
	}

	hashMap := make(map[int64]*st.UserProfileInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//查找本道题目最新回答的两个人
	sql := fmt.Sprintf(`select ownerUid ,answerId from v2answers where qid = %d and answerStatus = 0 order by answerTs desc limit 6`, qid)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()
	answerIds := make([]int, 0)
	for rows.Next() {
		var uid int64
		var answerId int
		rows.Scan(&uid, &answerId)
		if uid > 0 {
			answerIds = append(answerIds, answerId)

			//如果外漏回答的人  在正在参与讨论列表中  过滤
			//if uid == bestAnswer.OwnerInfo.Uin {
			//	continue
			//}

			//2018-08-01 头像去重按照answerId 不是uin
			if answerId == bestAnswer.AnswerId {
				continue
			}

			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			//hash 存储 userinfo
			//if _, ok := hashMap[uid]; ok {
			//	continue
			//}
			responders = append(responders, ui)
			//hashMap[uid] = ui
		}
	}

	count := len(responders)

	if count > 5 {
		responders = responders[0:5]
		return
	}

	//把所有answerId保存下来  不足五个人用评论者补上

	if len(answerIds) > 0 {
		GetCommentsByAnswerIds(qid, answerIds, hashMap, 5, &responders)
	}

	log.Debugf("qid :%d answerIds:%+v ", qid, answerIds)

	return
}

func GetCommentsByAnswerIds(qid int, answerIds []int, hashMap map[int64]*st.UserProfileInfo, restCnt int, responders *[]*st.UserProfileInfo) (err error) {

	str := ""
	for i, uid := range answerIds {
		if i != len(answerIds)-1 {
			str += fmt.Sprintf(`%d,`, uid)
		} else {
			str += fmt.Sprintf(`%d`, uid)
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//评论uid
	sql := fmt.Sprintf(`select ownerUid ,commentId from v2comments where answerId in ( %s ) `, str)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	commentIds := make([]int, 0)

	for rows.Next() {
		var uid int64
		var commentId int
		rows.Scan(&uid, &commentId)

		//		log.Debugf("comment uid:%d",uid)
		if uid > 0 {
			commentIds = append(commentIds, commentId)
			//hash 存储 userinfo
			//if _, ok := hashMap[uid]; ok {
			//	continue
			//}
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			*responders = append(*responders, ui)

			if len(*responders) == restCnt {
				return
			}
			hashMap[uid] = ui
		}
	}

	if len(commentIds) > 0 {
		GetReplysByCommentIds(commentIds, hashMap, restCnt, responders)
	}

	return
}

func GetReplysByCommentIds(commentIds []int, hashMap map[int64]*st.UserProfileInfo, restCnt int, responders *[]*st.UserProfileInfo) (err error) {

	str := ""
	for i, uid := range commentIds {
		if i != len(commentIds)-1 {
			str += fmt.Sprintf(`%d,`, uid)
		} else {
			str += fmt.Sprintf(`%d`, uid)
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//评论uid
	sql := fmt.Sprintf(`select fromUid from v2replys where commentId  in (%s)  group by fromUid`, str)

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
			//if _, ok := hashMap[uid]; ok {
			//	continue
			//}
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}
			*responders = append(*responders, ui)
			hashMap[uid] = ui
			if len(*responders) == restCnt {
				return
			}
		}
	}

	return
}

//参与讨论的人数 回答 评论 回复 回复的回复
func GetDiscussCnt(qid int) (discussCnt int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//评论uid
	sql := fmt.Sprintf(`select answerId  from v2answers where qid = %d and answerStatus = 0`, qid)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	answerIds := make([]int, 0)
	answerCnt := 0
	for rows.Next() {
		var answerId int
		rows.Scan(&answerId)
		answerIds = append(answerIds, answerId)
		answerCnt++
	}

	//log.Debugf("answerIds:%+v , answerCnt:%d",answerIds,answerCnt)

	if answerCnt == 0 {
		return
	}

	discussCnt = discussCnt + answerCnt

	commentIds, commentCnt, err := GetCntByType("commentId", "v2comments", "answerId", "commentStatus", answerIds)
	if err != nil {
		log.Errorf("get comment Info error !")
		return
	}
	if commentCnt == 0 {
		return
	}
	discussCnt = discussCnt + commentCnt

	replyIds, replyCnt, err := GetCntByType("replyId", "v2replys", "commentId", "replystatus", commentIds)
	if err != nil {
		log.Errorf("get reply Info error !")
		return
	}
	if replyCnt == 0 {
		return
	}
	discussCnt = discussCnt + replyCnt

	log.Debugf("answerCnt : %d  commentCnt : %d replyCnt : %d replyids :%v", answerCnt, commentCnt, replyCnt, replyIds)

	return

}

func GetCntByType(typeName string, table string, upperId string, typeStatus string, queryArr []int) (ids []int, idCnt int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//查询回复的总数
	str := ""
	for i, id := range queryArr {
		if i != len(queryArr)-1 {
			str += fmt.Sprintf(`%d,`, id)
		} else {
			str += fmt.Sprintf(`%d`, id)
		}
	}

	sql := fmt.Sprintf(`select %s  from %s where %s in (%s) and %s = 0 `, typeName, table, upperId, str, typeStatus)

	//log.Debugf("%s sql:%s",typeName ,sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	ids = make([]int, 0)
	idCnt = 0
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
		idCnt++
	}
	return
}
