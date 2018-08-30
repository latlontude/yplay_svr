package question

import (
	"api/board"
	"api/label"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"sort"
	"svr/st"
	"time"
)

type GetV2QuestionsformeReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

//提问总数 回答总数
type GetV2QuestionsRsp struct {
	V2Questions []*st.V2QuestionInfo `json:"questions"`
	TotalCnt    int                  `json:"totalCnt"`
	QuestionCnt int                  `json:"questionCnt"`
	AnswerCnt   int                  `json:"answerCnt"`
	LabelList   []*label.LabelInfo   `json:"labelList"`
	LoginDays   int                  `json:"loginDays"`
	LikeCnt     int                  `json:"likeCnt"`
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

func doGetV2QuestionsForMe(req *GetV2QuestionsformeReq, r *http.Request) (rsp *GetV2QuestionsRsp, err error) {

	log.Debugf("uin %d, GetQuestionsReq %+v", req.Uin, req)

	//我提出的问题
	questions, totalCnt, qstCnt, answerCnt, likeTotalCnt, err := GetV2QuestionsAndAnswer(req.Uin, 0, req.PageSize, req.PageNum)

	labelList, err := GetHomeLabelInfo(req.Uin, 0)

	if err != nil {
		log.Errorf("uin %d, doGetV2QuestionsForMe error, %s", req.Uin, err.Error())
		return
	}

	loginDays, err := GetRegisterTime(req.Uin)

	if err != nil {
		log.Errorf("get profile register ts error", req.Uin, err.Error())
	}

	rsp = &GetV2QuestionsRsp{questions, totalCnt, qstCnt, answerCnt, labelList, loginDays, likeTotalCnt}

	log.Debugf("uin %d, doGetV2QuestionsForMe success, rsp:%+v", req.Uin, rsp)

	return
}

/**
uin         自己uin
fuin        好友uin
pageSize    分页大小
pageNum     第几页
*/
func GetV2QuestionsAndAnswer(uin int64, fUin int64, pageSize int, pageNum int) (
	questions []*st.V2QuestionInfo, totalCnt int, qstCnt int, answerCnt int, likeTotalCnt int, err error) {

	questions = make([]*st.V2QuestionInfo, 0)
	likeTotalCnt = 0
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)

	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//直接查我回答的所有问题  answerStatus=1 代表删除

	var sql string

	//查询好友的个人主页 看不到匿名问题(回答不需要判断匿名问题)
	if fUin > 0 {
		sql = fmt.Sprintf(`select v2answers.answerId,v2answers.qid,v2answers.answerContent,v2answers.answerImgUrls,v2answers.answerTs,v2answers.ext,
v2questions.qid, v2questions.boardId,v2questions.ownerUid,v2questions.qTitle,v2questions.qContent,
v2questions.qImgUrls,v2questions.qType,v2questions.isAnonymous,v2questions.createTs,v2questions.modTs,v2questions.sameAskUid,v2questions.ext
from   v2answers ,v2questions
where v2answers.answerStatus = 0 
and v2questions.qStatus = 0 
and v2answers.qid=v2questions.qid
and v2questions.isAnonymous = 0
and v2answers.ownerUid = %d`, fUin)
	} else {
		//自己看自己主页 可以看到匿名问题
		sql = fmt.Sprintf(`select v2answers.answerId,v2answers.qid,v2answers.answerContent,v2answers.answerImgUrls,v2answers.answerTs,v2answers.ext,
v2questions.qid, v2questions.boardId,v2questions.ownerUid,v2questions.qTitle,v2questions.qContent,
v2questions.qImgUrls,v2questions.qType,v2questions.isAnonymous,v2questions.createTs,v2questions.modTs,v2questions.sameAskUid,v2questions.ext
from   v2answers ,v2questions
where v2answers.answerStatus = 0 
and v2questions.qStatus = 0 
and v2answers.qid=v2questions.qid 
and v2answers.ownerUid = %d`, uin)
	}

	log.Debugf("sql:%s", sql)

	rows, err := inst.Query(sql)
	defer rows.Close()

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	answerCnt = 0

	for rows.Next() {

		var info st.V2QuestionInfo    //问题
		var answerInfo st.AnswersInfo //回答
		var boardInfo st.BoardInfo    //墙信息
		var uid int64

		var boardId int

		var sameAskUid string
		rows.Scan(
			&answerInfo.AnswerId,
			&answerInfo.Qid,
			&answerInfo.AnswerContent,
			&answerInfo.AnswerImgUrls,
			&answerInfo.AnswerTs,
			&answerInfo.Ext,
			&info.Qid,
			&boardId,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.QType,
			&info.IsAnonymous,
			&info.CreateTs,
			&info.ModTs,
			&sameAskUid,
			&info.Ext)

		boardInfo, err = board.GetBoardInfoByBoardId(uin, boardId)
		info.Board = &boardInfo

		//回答问题的时间比问题创建时间更新  统一用createTs排序
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
		likeTotalCnt = likeTotalCnt + likeCnt

		answerInfo.LikeCnt = likeCnt
		isILike, err1 := checkIsILikeAnswer(uin, answerInfo.AnswerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		answerInfo.IsILike = isILike
		info.BestAnswer = &answerInfo
		questions = append(questions, &info)

		answerCnt = answerCnt + 1
	}

	//查询所有我提出的问题

	if fUin > 0 {
		sql = fmt.Sprintf(`select * from v2questions 
				where v2questions.qStatus = 0 
				and v2questions.isAnonymous = 0
				and v2questions.ownerUid = %d`, fUin)
	} else {
		sql = fmt.Sprintf(`select * from v2questions
				where v2questions.qStatus = 0 
				and v2questions.ownerUid = %d`, uin)
	}

	rows, err = inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}

	qstCnt = 0
	for rows.Next() {
		var info st.V2QuestionInfo
		var uid int64
		var boardInfo st.BoardInfo
		var boardId int
		var qStatus int64
		var sameAskUid string

		rows.Scan(
			&info.Qid,
			&boardId,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.QType,
			&info.IsAnonymous,
			&qStatus,
			&info.CreateTs,
			&info.ModTs,
			&sameAskUid,
			&info.Ext,
		)

		boardInfo, err = board.GetBoardInfoByBoardId(uin, boardId)
		info.Board = &boardInfo
		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		answerCnt, err := GetAnswerCnt(info.Qid)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		info.AnswerCnt = answerCnt
		questions = append(questions, &info)
		qstCnt = qstCnt + 1
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
	totalCnt = len(questions)
	if e > totalCnt {
		e = totalCnt
	}

	//s - e
	questions = questions[s:e]
	log.Debugf("end GetV2QuestionsAndAnswer uin:%d,fuid:%d,TotalCnt:%d qstCnt:%d,answerCnt:%d",
		uin, fUin, totalCnt, qstCnt, answerCnt)
	return
}

func GetHomeLabelInfo(uin int64, fUin int64) (labelList []*label.LabelInfo, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	var uid int64
	if fUin > 0 {
		uid = fUin
	} else {
		uid = uin
	}

	sql := fmt.Sprintf(`select experience_label.labelId ,experience_label.labelName ,experience_label.boardId from experience_label
	where experience_label.labelId in
	(select experience_share.labelId from experience_share
		where answerId in (select answerId from v2answers where ownerUid = %d )
	     group by experience_share.labelId)`, uid)

	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var labelInfo label.LabelInfo
		rows.Scan(&labelInfo.LabelId, &labelInfo.LabelName, &labelInfo.BoardId)
		labelList = append(labelList, &labelInfo)
	}

	return
}

func GetRegisterTime(uin int64) (loginDays int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select ts from profiles where uin = %d`, uin)

	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var registerTime int64
	for rows.Next() {
		rows.Scan(&registerTime)
	}

	now := time.Now().Unix()

	loginDays = int((now - registerTime) / 86400)

	return

}
