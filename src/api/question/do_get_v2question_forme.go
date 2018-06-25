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
	mapAnswer := make(map[int]int)
	var sql = fmt.Sprintf(`select qid, answerTs from  v2answers where ownerUid = %d order by answerTs desc`, uin)

	rows, err := inst.Query(sql)
	defer rows.Close()

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	//映射  qid ---> ansewrTs
	for rows.Next() {
		var qid int
		var answerTs int
		rows.Scan(&qid, &answerTs)
		mapAnswer[qid] = answerTs
	}

	sql = fmt.Sprintf(`select count(qid) as cnt from  v2questions where qStatus = 0 and ownerUid = %d 
								or  qid in (select qid from v2answers where ownerUid = %d) 
								order by createTs desc`, uin, uin)

	rows, err = inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		rows.Scan(&TotalCnt)
	}

	if TotalCnt == 0 {
		return
	}

	sql = fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls, isAnonymous, createTs, modTs 
								from v2questions where qStatus = 0 and ownerUid = %d 
								or  qid in (select qid from v2answers where ownerUid = %d) order by createTs desc `, uin, uin)

	rows, err = inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}

	for rows.Next() {
		var info st.V2QuestionInfo
		var uid int64

		rows.Scan(&info.Qid, &uid, &info.QTitle, &info.QContent, &info.QImgUrls, &info.IsAnonymous, &info.CreateTs, &info.ModTs)

		if v, ok := mapAnswer[info.Qid]; ok {
			//如果answerTs比createTs大 更新createTs
			if v > info.CreateTs {
				info.CreateTs = v
			}
		}

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
	e := pageSize
	questions = questions[s : s+e]

	log.Debugf("end GetV2Questionsforme uin:%d TotalCnt:%d", uin, TotalCnt)
	return
}
