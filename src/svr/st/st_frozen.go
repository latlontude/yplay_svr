package st

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"strconv"
	"time"
)

func GetFreezingStatus(uin int64) (freezeTs int, err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	sql := fmt.Sprintf(`select freezeTs from freezingStatus where uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		rows.Scan(&freezeTs)

		find = true
	}

	if !find {
		freezeTs = 0
	}

	return
}

func EnterFrozenStatus(uin int64) (err error) {

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	ts := int(time.Now().Unix())

	sql := fmt.Sprintf(`insert into freezingStatus values(%d, %d, %d) on duplicate key update freezeTs = %d, ts = %d`, uin, ts, ts, ts, ts)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

func LeaveFrozenStatusByInviteFriend(uin int64) (err error) {

	/*
		app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_VOTE_PROGRESS)
		if err != nil {
			log.Error(err.Error())
			return
		}

		keyStr := fmt.Sprintf("%d", uin)

		left, err := app.ZCard(keyStr)
		if err != nil {
			log.Error(err.Error())
			return
		}

		//当前有未回答完成的，则不改变状态
		if left > 0 {
			return
		}
	*/

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	log.Debugf("uin %d leave frozenstatus", uin)

	ts := int(time.Now().Unix())

	sql := fmt.Sprintf(`update freezingStatus set freezeTs = 0, ts = %d where uin = %d`, ts, uin)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

//答题或者投票都会更新冷冻状态 如果是最后一道题 则进入冷冻状态
func UpdateVoteProgress2(uin int64, qid int, index int) (err error) {

	log.Debugf("UpdateCurrentVoteProgress2 uin %d, qid %d, index %d", uin, qid, index)

	if uin == 0 || qid == 0 || index <= 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_QINFO)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"qid", "qindex"}

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(valsStr) != len(fields) || len(valsStr) == 0 {
		err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
		log.Errorf(err.Error())
		return
	}

	lastQId, _ := strconv.Atoi(valsStr["qid"])
	lastQIndex, _ := strconv.Atoi(valsStr["qindex"])

	if lastQId != qid || lastQIndex != index {
		err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
		log.Errorf(err.Error())
		return
	}

	//如果是最后一题，强制跟客户端同步进入冷冻状态
	if index == constant.ENUM_QUESTION_BATCH_SIZE {

		log.Errorf("uin %d enter frozenstatus", uin)

		err = EnterFrozenStatus(uin)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	kvs := make(map[string]string)
	kvs["voted"] = "1"

	err = app.HMSet(keyStr, kvs)
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}

//答题或者投票都会更新冷冻状态 如果是最后一道题 则进入冷冻状态
func UpdateVoteProgressByPreGene(uin int64, qid int, index int) (err error) {

	log.Debugf("UpdateVoteProgressByPreGene uin %d, qid %d, index %d", uin, qid, index)

	if uin == 0 || qid == 0 || index <= 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_progress", uin)

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"qid", "qindex"}

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(valsStr) != len(fields) || len(valsStr) == 0 {
		err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
		log.Errorf(err.Error())
		return
	}

	lastQId, _ := strconv.Atoi(valsStr["qid"])
	lastQIndex, _ := strconv.Atoi(valsStr["qindex"])

	//第一次切换时，可能和当前题目对应不上
	if lastQId != 0 {
		if lastQId != qid || lastQIndex != index {
			err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
			log.Errorf(err.Error())
			return
		}
	}

	//如果是最后一题，强制跟客户端同步进入冷冻状态
	if index == constant.ENUM_QUESTION_BATCH_SIZE {

		log.Errorf("uin %d enter frozenstatus", uin)

		err = EnterFrozenStatus(uin)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	kvs := make(map[string]string)
	kvs["voted"] = "1"

	err = app.HMSet(keyStr, kvs)
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
