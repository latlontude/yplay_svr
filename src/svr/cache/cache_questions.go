package cache

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func CacheQuestions() (err error) {

	QUESTIONS = make(map[int]*st.QuestionInfo)

	ALL_GENE_QIDS = make([]int, 0)
	ALL_BOY_QIDS = make([]int, 0)
	ALL_GIRL_QIDS = make([]int, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select qid, qtext, qiconUrl,  optionGender, replyGender, schoolType, dataSrc, Delivery, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts from questions2 order by qid`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.QuestionInfo
		rows.Scan(&info.QId, &info.QText, &info.QIconUrl, &info.OptionGender, &info.ReplyGender, &info.SchoolType, &info.DataSrc, &info.Delivery, &info.Status, &info.TagId, &info.TagName,
			&info.SubTagId1, &info.SubTagName1, &info.SubTagId2, &info.SubTagName2, &info.SubTagId3, &info.SubTagName3, &info.Ts)

		info.QIconUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/qicon/%s", info.QIconUrl)

		QUESTIONS[info.QId] = &info

		//不在被选中的问题列表中
		//历史上可能是选中问题列表历史用户答过这道题目
		if info.Status > 0 {
			continue
		}

		if info.OptionGender == 0 {
			ALL_GENE_QIDS = append(ALL_GENE_QIDS, info.QId)
		} else if info.OptionGender == 1 {
			ALL_BOY_QIDS = append(ALL_BOY_QIDS, info.QId)
		} else {
			ALL_GIRL_QIDS = append(ALL_GIRL_QIDS, info.QId)
		}
	}

	return
}

func AddCacheQuestions(qid int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select qid, qtext, qiconUrl,  optionGender, replyGender, schoolType, dataSrc, delivery, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts from questions2 where qid = %d`, qid)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.QuestionInfo
		rows.Scan(&info.QId, &info.QText, &info.QIconUrl, &info.OptionGender, &info.ReplyGender, &info.SchoolType, &info.DataSrc, &info.Delivery, &info.Status, &info.TagId, &info.TagName,
			&info.SubTagId1, &info.SubTagName1, &info.SubTagId2, &info.SubTagName2, &info.SubTagId3, &info.SubTagName3, &info.Ts)

		info.QIconUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/qicon/%s", info.QIconUrl)

		QUESTIONS[info.QId] = &info

		//不在被选中的问题列表中
		//历史上可能是选中问题列表历史用户答过这道题目
		if info.Status > 0 {
			continue
		}

		if info.OptionGender == 0 {
			ALL_GENE_QIDS = append(ALL_GENE_QIDS, info.QId)
		} else if info.OptionGender == 1 {
			ALL_BOY_QIDS = append(ALL_BOY_QIDS, info.QId)
		} else {
			ALL_GIRL_QIDS = append(ALL_GIRL_QIDS, info.QId)
		}
	}

	return
}
