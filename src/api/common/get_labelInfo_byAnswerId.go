package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func GetLabelInfoByAnswerId(answerId int) (expLabels []*st.ExpLabel, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select experience_label.labelId, experience_label.labelName from experience_share,experience_label 
where experience_share.labelId = experience_label.labelId and experience_share.answerId = %d`, answerId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var expLabel st.ExpLabel
		rows.Scan(&expLabel.LabelId, &expLabel.LabelName)
		expLabels = append(expLabels, &expLabel)
	}

	return
}
