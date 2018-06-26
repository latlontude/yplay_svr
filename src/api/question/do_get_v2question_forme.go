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


	//直接查我回答的所有问题  answerStatus=1 代表删除
	var sql = fmt.Sprintf(`select * from  v2answers ,v2questions ,v2boards
		where v2answers.answerStatus = 0 
		and v2questions.qStatus = 0 
		and v2answers.qid=v2questions.qid 
		and v2questions.boardId = v2boards.boardId
		and v2answers.ownerUid = %d`,uin)

	rows, err := inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var info st.V2QuestionInfo		//问题
		var answerInfo st.AnswersInfo	//回答
		var boardInfo  st.BoardInfo		//墙信息
		var uid int64


		var OwnerInfo st.UserProfileInfo	//墙主info
		var answerStatus string
		var boardId int64
		var qStatus int64

		var temp string					//暂时不需要给前段的字段 用temp接收

		rows.Scan(
			&answerInfo.AnswerId,
			&answerInfo.Qid,
			&answerInfo.AnswerContent,
			&answerInfo.AnswerImgUrls,
			&uid,
			&answerStatus,
			&answerInfo.AnswerTs,
			&info.Qid,
			&boardId,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.IsAnonymous,
			&qStatus,
			&info.CreateTs,
			&info.ModTs,
			&boardInfo.BoardId,
			&temp,
			&temp,
			&temp,
			&temp,
			&temp,
			&OwnerInfo.Uin,			//取墙主uid
			&temp,
			)

		boardInfo.OwnerInfo = &OwnerInfo
		info.Board = &boardInfo
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


	//查询所有我提出的问题
	sql = fmt.Sprintf(`select * from v2questions ,v2boards where v2questions.qStatus = 0 
				and v2questions.boardId=v2boards.boardId 
				and v2questions.ownerUid = %d`, uin)

	rows, err = inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}

	for rows.Next() {
		var info st.V2QuestionInfo
		var uid int64

		var boardInfo  st.BoardInfo

		var boardId int64
		var qStatus int64

		var temp string
		var OwnerInfo st.UserProfileInfo

		rows.Scan(
			&info.Qid,
			&boardId,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.IsAnonymous,
			&qStatus,
			&info.CreateTs,
			&info.ModTs,
			&boardInfo.BoardId,
			&temp,
			&temp,
			&temp,
			&temp,
			&temp,
			&OwnerInfo.Uin,
			&temp,
			)

		boardInfo.OwnerInfo = &OwnerInfo
		info.Board = &boardInfo
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

	//s - e
	questions = questions[s : e]
	log.Debugf("end GetV2QuestionsForMe uin:%d TotalCnt:%d", uin, TotalCnt)
	return
}
