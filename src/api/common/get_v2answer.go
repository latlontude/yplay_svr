package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func GetV2Answer(answerId int) (answer st.AnswersInfo, err error) {
	if answerId == 0 {
		log.Errorf("answerId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qid, answerId, ownerUid, answerContent, answerImgUrls, isAnonymous,answerTs from  v2answers where answerId = %d`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var uid int64
	for rows.Next() {
		rows.Scan(&answer.Qid, &answer.AnswerId, &uid, &answer.AnswerContent, &answer.AnswerImgUrls, &answer.IsAnonymous, &answer.AnswerTs)
	}

	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		answer.OwnerInfo = ui
	}

	//查找该问题的labelName
	expLabels, err3 := GetLabelInfoByAnswerId(answer.AnswerId)
	if err3 == nil {
		answer.ExpLabel = expLabels
	}

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
