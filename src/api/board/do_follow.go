package board

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type FollowReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int64 `schema:"boardId"`
}

type FollowRsp struct {
	Code int `json:"code"` // 0 表示成功
}

func doFollow(req *FollowReq, r *http.Request) (rsp *FollowRsp, err error) {

	log.Debugf("uin %d, FollowReq %+v", req.Uin, req)

	code, err := Follow(req.Uin, req.BoardId)

	if err != nil {
		log.Errorf("uin %d, Follow error, %s", req.Uin, err.Error())
		return
	}

	rsp = &FollowRsp{code}

	log.Debugf("uin %d, FollowRsp succ, %+v", req.Uin, rsp)

	return
}

func Follow(uin, boardId int64) (code int, err error) {
	log.Debugf("start Follow uin:%d boardId:%d", uin, boardId)

	code = -1
	if uin <= 0 || boardId <= 0 {
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

	sql := fmt.Sprintf(`select uin from v2follow where uin = %d and boardId = %d and status = 0`, uin, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rest.NewAPIError(constant.E_REPORT_TYPE_REPEAT, "repeat follow")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	stmt, err := inst.Prepare(`insert into v2follow values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, boardId, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	code = 0
	log.Debugf("end Follow uin:%d code:%d", uin, code)
	return
}
