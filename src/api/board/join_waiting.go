package board

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type JoinReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SchoolId int    `schema:"schoolId"`
	BoardId  int    `schema:"boardId"`
	IsJoin   bool   `schema:"isJoin"`
}

type JoinRsp struct {
	TotalCnt int  `json:"totalCnt"` //等待天使人数
	HasJoin  bool `json:"hasJoin"`
}

func doJoin(req *JoinReq, r *http.Request) (rsp *JoinRsp, err error) {

	log.Debugf("uin %d, FollowReq %+v", req.Uin, req)

	//boardId ,err := GetBoardIdBySchoolId(req.SchoolId)

	totalCnt, hasJoin, err := Join(req.Uin, req.SchoolId, req.IsJoin)

	if err != nil {
		log.Errorf("uin %d, Follow error, %s", req.Uin, err.Error())
		return
	}

	rsp = &JoinRsp{totalCnt, hasJoin}

	log.Debugf("uin %d, FollowRsp succ, %+v", req.Uin, rsp)

	return
}

func Join(uin int64, schoolId int, isJoin bool) (totalCnt int, hasJoin bool, err error) {

	totalCnt, hasJoin, err = GetJoinCnt(uin, schoolId)

	log.Debugf("join cnt :%d", totalCnt)

	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if isJoin {
		err = InsertJoin(uin, schoolId)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		totalCnt++
	}

	log.Debugf("end join uin:%d totalCnt:%d", uin, totalCnt)

	return
}

func GetJoinCnt(uin int64, schoolId int) (totalCnt int, hasJoin bool, err error) {
	totalCnt = 0

	if uin <= 0 || schoolId <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select uin from join_waiting_angel where schoolId = %d`, schoolId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	hasJoin = false
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)

		if uin == uid {
			hasJoin = true
		}
		totalCnt++
	}
	return
}

func InsertJoin(uin int64, schoolId int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(uin) from join_waiting_angel where uin = %d and schoolId = %d`, uin, schoolId)

	log.Debugf("sql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		rows.Scan(&count)
	}

	log.Debugf("count：%d", count)
	if count > 0 {
		err = rest.NewAPIError(constant.E_REPORT_TYPE_REPEAT, "repeat join")
		log.Errorf(err.Error())
		return
	}

	stmt, err := inst.Prepare(`insert into join_waiting_angel values(?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	_, err = stmt.Exec(0, uin, schoolId, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
