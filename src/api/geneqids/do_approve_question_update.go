package geneqids

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"time"
)

type UserSimpleInfo struct {
	Uin        int64 `json:"uin"`
	Gender     int   `json:"gender"`
	Grade      int   `json:"grade"`
	SchoolId   int   `json:"schoolId"`
	SchoolType int   `json:"schoolType"`
}

type ApproveQuestionUpdateReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	User int64 `schema:"user"`
	QId  int   `schema:"qid"`
}

type ApproveQuestionUpdateRsp struct {
	Pos int `json:"pos"`
}

func doApproveQuestionUpdate(req *ApproveQuestionUpdateReq, r *http.Request) (rsp *ApproveQuestionUpdateRsp, err error) {

	log.Errorf("uin %d, ApproveQuestionUpdateReq %+v", req.Uin, req)

	pos, err := InsertApprovedQId(req.User, req.QId)
	if err != nil {
		log.Errorf("uin %d, ApproveQuestionUpdateRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &ApproveQuestionUpdateRsp{pos}

	log.Errorf("uin %d, ApproveQuestionUpdateRsp succ, %+v", req.Uin, rsp)

	return
}

func ApproveQuestionUpdate(uin int64, qid int) (err error) {

	if uin == 0 || qid == 0 {
		return
	}

	uins, err := GetSameSchoolGradeUins(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rand.Seed(time.Now().Unix())

	for _, uid := range uins {
		_, err = InsertApprovedQId(uid, qid)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
	}

	return
}

func GetSameSchoolGradeUins(uin int64) (uins []int64, err error) {

	uins = make([]int64, 0)

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, gender, grade, schoolId, schoolType from profiles`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	ums := make(map[int64]*UserSimpleInfo)

	for rows.Next() {

		var usi UserSimpleInfo
		rows.Scan(&usi.Uin, &usi.Gender, &usi.Grade, &usi.SchoolId, &usi.SchoolType)

		ums[usi.Uin] = &usi
	}

	var ui *UserSimpleInfo
	if _, ok := ums[uin]; !ok {
		log.Errorf("GetSameSchoolGradeUins uin %d, not exist!", uin)
		return
	}

	ui = ums[uin]

	for uid, usi := range ums {
		//同校同年级的
		if usi.SchoolId == ui.SchoolId && usi.Grade == ui.Grade {
			uins = append(uins, uid)
		}
	}

	return
}

func InsertApprovedQId(uin int64, qid int) (pos int, err error) {

	if uin == 0 || qid == 0 {
		return
	}

	log.Debugf("uin %d, begin InsertApprovedQId qid %d", uin, qid)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"cursor", "insertcursor"}

	keyStr := fmt.Sprintf("%d_progress", uin)
	keyStr2 := fmt.Sprintf("%d_qids", uin)

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, InsertApprovedQId HMGet rsp %+v", uin, valsStr)

	if _, ok := valsStr["cursor"]; !ok {
		err = rest.NewAPIError(constant.E_PRE_GENE_QIDS_PROGRESS_ERR, "pre gene qids progress info error")
		log.Errorf(err.Error())
		return
	}

	orgPos := -1
	//如果从来没有答题 则上一次题目设置为0  上答题一次索引为0
	if len(valsStr["cursor"]) > 0 {
		orgPos, _ = strconv.Atoi(valsStr["cursor"])
	}

	insertcursor := -1
	if len(valsStr["insertcursor"]) > 0 {
		insertcursor, _ = strconv.Atoi(valsStr["insertcursor"])
	}

	if insertcursor <= 0 {
		//insertcursor从来没有设置过
		pos = orgPos + 1 + rand.Intn(3)

	} else {

		if insertcursor > orgPos {
			//插入位置始终比当前答题的进度要快一些
			pos = insertcursor + 1 + rand.Intn(3)
		} else if orgPos-insertcursor > 100 {
			//插入比答题快很多，并且回绕的情况
			pos = insertcursor + 1 + rand.Intn(3)
		} else {
			//插入慢，答题快
			pos = orgPos + 1 + rand.Intn(3)
		}
	}

	total, err := app.ZCard(keyStr2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if total == 0 {
		log.Errorf("uin %d, qid %d, InsertApprovedQId totalCnt 0", uin, qid)
		return
	}

	if pos >= total {
		pos = pos % total
	}

	log.Debugf("uin %d, InsertApprovedQId qid %d, total %d, orgcursor %d, insertcursor, newPos %d", uin, qid, total, orgPos, insertcursor, pos)

	err = app.ZAdd(keyStr2, int64(pos), fmt.Sprintf("%d", qid))
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//更新上次的插入进度
	res := make(map[string]string)
	res["insertcursor"] = fmt.Sprintf("%d", insertcursor)

	err1 := app.HMSet(keyStr, res)
	if err1 != nil {
		log.Errorf(err1.Error())
	}

	return
}
