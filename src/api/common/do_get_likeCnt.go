package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

func GetAnswerLikeCnt(answerId int) (cnt int, err error) {
	//log.Debugf("start getAnswerLikeCnt answerId:%d", answerId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	// 点赞数
	sql := fmt.Sprintf(`select count(id) as cnt from v2likes where type = 1 and likeId = %d and likeStatus != 2`, answerId)
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

	log.Debugf("end getAnswerLikeCnt answerId:%d cnt:%d", answerId, cnt)
	return
}

func CheckIsILikeAnswer(uin int64, answerId int) (ret bool, err error) {
	//log.Debugf("start checkIsILikeAnswer uin:%d answerId:%d", uin, answerId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select id from v2likes where type = 1 and likeId = %d and ownerUid = %d and likeStatus != 2`, answerId, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		rows.Scan(&id)
		ret = true
	}

	//log.Debugf("end checkIsILikeAnswer uin:%d answerId:%d ret:%t", uin, answerId, ret)
	return
}
