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

type GetV2QuestionsformeReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

// 自定义排序
type quest []*st.V2QuestionInfo

func (I quest) Len() int {
	return len(I)
}

func (I quest) Less(i, j int) bool {
	return I[i].CreateTs > I[j].CreateTs
}

func (I quest) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

func doGetv2Questionsforme(req *GetV2QuestionsformeReq, r *http.Request) (rsp *GetQuestionsRsp, err error) {

	log.Debugf("uin %d, GetQuestionsReq %+v", req.Uin, req)

	//我提出的问题
	questions, TotalCnt, err := GetV2QuestionsAndAnswer(req.Uin, req.PageSize, req.PageNum)

	if err != nil {
		log.Errorf("uin %d, GetV2Questionsforme error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetQuestionsRsp{questions, TotalCnt}

	log.Debugf("uin %d, doGetv2Questionsforme succ, %+v", req.Uin, rsp)

	return
}

func GetV2QuestionsAndAnswer(uin int64, pageSize int, pageNum int) (questions []*st.V2QuestionInfo, TotalCnt int, err error) {

	questions = make([]*st.V2QuestionInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)

	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//查询answers表 找到回答问题的qid  answerTs
	//mapAnswer := make(map[int]int)
	//var sql = fmt.Sprintf(`select qid, answerTs from  v2answers where ownerUid = %d order by answerTs desc`, uin)

	//直接查我回答的所有问题
	var sql = fmt.Sprintf(`select * from  v2answers ,v2questions where v2answers.ownerUid=v2questions.ownerUid and v2answers.qid=v2questions.qid and  v2questions.ownerUid =%d`, uin)

	rows, err := inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var info st.V2QuestionInfo
		var answerInfo st.AnswersInfo
		var uid int64
		var temp string

		rows.Scan(
			&answerInfo.AnswerId,
			&answerInfo.Qid,
			&answerInfo.AnswerContent,
			&answerInfo.AnswerImgUrls,
			&uid,
			&temp,
			&answerInfo.AnswerTs,
			&info.Qid,
			&temp,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.IsAnonymous,
			&temp,
			&info.CreateTs,
			&info.ModTs)

		info.CreateTs = answerInfo.AnswerTs
		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}
			info.OwnerInfo = ui
		}

		commentCnt, err1 := getCommentCnt(answerInfo.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		answerInfo.CommentCnt = commentCnt

		//点赞数
		likeCnt, err1 := getAnswerLikeCnt(answerInfo.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}
		answerInfo.LikeCnt = likeCnt
		isILike, err1 := checkIsILikeAnswer(uin, answerInfo.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		answerInfo.IsILike = isILike
		info.BestAnswer = &answerInfo
		questions = append(questions, &info)
	}

	sql = fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls, isAnonymous, createTs, modTs 
								from v2questions where qStatus = 0 and ownerUid = %d`, uin)
	rows, err = inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}

	for rows.Next() {
		var info st.V2QuestionInfo
		var uid int64

		rows.Scan(&info.Qid, &uid, &info.QTitle, &info.QContent, &info.QImgUrls, &info.IsAnonymous, &info.CreateTs, &info.ModTs)

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
		questions = append(questions, &info)
	}

	//排序
	sort.Sort(quest(questions))

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := s + pageSize

	// 超过最大长度  等于slice长度
	TotalCnt = len(questions)

	if e > TotalCnt {
		e = TotalCnt
	}
	questions = questions[s : s+e]
	log.Debugf("end GetV2Questionsforme uin:%d TotalCnt:%d", uin, TotalCnt)
	return
}
