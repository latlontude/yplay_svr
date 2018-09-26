package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetQuestionsDetailReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int `schema:"boardId"`
	Qid     int `schema:"qid"`
	Version int `schema:"version"`
}

type GetQuestionsDetailRsp struct {
	Questions st.V2QuestionInfo `json:"question"`
}

func doGetOneQuestionDetail(req *GetQuestionsDetailReq, r *http.Request) (rsp *GetQuestionsDetailRsp, err error) {

	log.Debugf("uin %d, GetOneQuestionsReq %+v", req.Uin, req)

	question, err := GetOneQuestionDetail(req.Uin, req.BoardId, req.Qid)

	if err != nil {
		log.Errorf("uin %d, GetQuestions error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetQuestionsDetailRsp{question}

	log.Debugf("uin %d, GetQuestionsRsp succ  , rsp:%v", req.Uin, rsp)

	return
}

func GetOneQuestionDetail(uin int64, boardId, qid int) (question st.V2QuestionInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qid, ownerUid, qTitle, qContent, qImgUrls,qType, isAnonymous, createTs, modTs ,ext,longitude,latitude, poiTag from v2questions 
		where qStatus = 0 and (qContent != "" or qImgUrls != "") and boardId = %d and qid = %d `, boardId, qid)

	rows, err := inst.Query(sql)
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
			&question.Ext,
			&question.Longitude,
			&question.Latitude,
			&question.PoiTag)

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
