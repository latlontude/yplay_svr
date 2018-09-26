package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func GetV2Question(qid int, version int) (question st.V2QuestionInfo, err error) {
	if qid == 0 {
		log.Errorf("qid is zero")
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "qid is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qid, qTitle, qContent, qImgUrls,qType, ownerUid, isAnonymous, createTs, modTs ,ext from  v2questions where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var ownerUid int64

	for rows.Next() {
		rows.Scan(&question.Qid, &question.QTitle, &question.QContent, &question.QImgUrls, &question.QType,
			&ownerUid, &question.IsAnonymous, &question.CreateTs, &question.ModTs, &question.Ext)
	}
	question.QContent = GetContentByVersion(question.QContent, question.QType, version)
	if ownerUid > 0 {
		ui, err1 := st.GetUserProfileInfo(ownerUid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		question.OwnerInfo = ui
	}
	answerCnt, err := GetAnswerCnt(qid)
	question.AnswerCnt = answerCnt
	return
}

func GetQuestionsByQidStr(qidStr string, version int) (questions []*st.V2QuestionInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qid, qTitle, qContent, qImgUrls,qType, ownerUid, isAnonymous, createTs, modTs ,ext,latitude,longtitude,poiTag
from  v2questions where qid = in (%s)`, qidStr)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var ownerUid int64

	for rows.Next() {
		var question st.V2QuestionInfo
		rows.Scan(&question.Qid, &question.QTitle, &question.QContent, &question.QImgUrls, &question.QType,
			&ownerUid, &question.IsAnonymous, &question.CreateTs, &question.ModTs, &question.Ext)
		question.QContent = GetContentByVersion(question.QContent, question.QType, version)

		if ownerUid > 0 {
			ui, err1 := st.GetUserProfileInfo(ownerUid)
			if err1 != nil {
				log.Error(err1.Error())
			}

			question.OwnerInfo = ui
		}
		answerCnt, _ := GetAnswerCnt(question.Qid)
		question.AnswerCnt = answerCnt

		questions = append(questions, &question)
	}
	return
}

func GetContentByVersion(qContent string, qType int, version int) (newContent string) {
	if qType >= 3 && version < 4989 {
		newContent = qContent + "【未识别此内容～请升级到最新版本！感恩～】"
	} else {
		newContent = qContent
	}
	return
}
