package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetOneQuestionsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int `schema:"boardId"`
	Index   int `schema:"index"`
	Version int `schema:"version"`
}

type GetOneQuestionsRsp struct {
	Questions st.V2QuestionInfo `json:"question"`
}

func doGetOneQuestions(req *GetOneQuestionsReq, r *http.Request) (rsp *GetOneQuestionsRsp, err error) {

	log.Debugf("uin %d, GetOneQuestionsReq %+v", req.Uin, req)

	question, err := GetOneQuestions(req.Uin, req.BoardId, req.Index)

	if err != nil {
		log.Errorf("uin %d, GetQuestions error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetOneQuestionsRsp{question}

	log.Debugf("uin %d, GetQuestionsRsp succ  , rsp:%v", req.Uin, rsp)

	return
}

func GetOneQuestions(uin int64, boardId, index int) (question st.V2QuestionInfo, err error) {

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

	var totalCnt int
	for rows.Next() {
		rows.Scan(&totalCnt)
	}

	if totalCnt == 0 {
		return
	} else {
		index = index % totalCnt
	}

	sql = fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls,qType, isAnonymous, createTs, modTs ,ext from v2questions 
		where qStatus = 0 and (qContent != "" or qImgUrls != "") and boardId = %d order by qid desc limit %d ,1`, boardId, index)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		rows.Scan(
			&question.Qid,
			&uid,
			&question.QTitle,
			&question.QContent,
			&question.QImgUrls,
			&question.QType,
			&question.IsAnonymous,
			&question.CreateTs,
			&question.ModTs,
			&question.Ext)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			question.OwnerInfo = ui
		}
		//answerCnt, err := GetAnswerCnt(info.Qid)
		answerCnt, err := GetDiscussCnt(question.Qid)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		question.AnswerCnt = answerCnt

		bestAnswer, _ := GetBestAnswer(uin, question.Qid)
		question.BestAnswer = bestAnswer
		if bestAnswer != nil {
			responders, _ := GetQidNewResponders(question.Qid, bestAnswer)
			question.NewResponders = responders
		}
	}
	return
}
