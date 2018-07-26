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

	sql := fmt.Sprintf(`select qid, answerId, ownerUid, answerContent, answerImgUrls, answerTs from  v2answers where answerId = %d`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var uid int64
	for rows.Next() {
		rows.Scan(&answer.Qid, &answer.AnswerId, &uid, &answer.AnswerContent, &answer.AnswerImgUrls, &answer.AnswerTs)
	}

	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		answer.OwnerInfo = ui
	}

	return
}
