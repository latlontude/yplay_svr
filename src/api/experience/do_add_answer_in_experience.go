package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type AddAnswerInExperienceReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	Qid      int    `schema:"qid"`
	AnswerId int    `schema:"answerId"`
	LabelId  int    `schema:"labelId"`
	BoardId  int    `schema:"boardId"`
}

type AddAnswerInExperienceRsp struct {
}

func doAddAnswerIdInExp(req *AddAnswerInExperienceReq, r *http.Request) (rsp *AddAnswerInExperienceRsp, err error) {

	log.Debugf("uin %d, AddAnswerInExperienceReq succ, %+v", req.Uin, rsp)

	err = AddAnswerIdInExp(req.Uin, req.BoardId, req.Qid, req.AnswerId, req.LabelId)

	if err != nil {
		log.Errorf("uin %d, AddAnswerInExperienceReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddAnswerInExperienceRsp{}
	return
}

func AddAnswerIdInExp(uin int64, boardId, qid, answerId, labelId int) (err error) {

	//校验权限
	hasPermission, err := CheckPermit(uin, boardId, labelId)

	if !hasPermission {
		err = rest.NewAPIError(constant.E_DB_QUERY, "add answer has not  permit")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into experience_share(id, boardId, labelId, qid,answerId, operator,ts, status) 
		values(?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	_, err = stmt.Exec(0, boardId, labelId, qid, answerId, uin, ts, 0)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

func CheckPermit(uin int64, boardId int, labelId int) (hasPermission bool, err error) {

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//board
	sql := fmt.Sprintf(`select ownerUid from v2boards where boardId = %d`, boardId)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	hasPermission = false

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		if uid == uin {
			hasPermission = true
		}
	}

	//admin
	sql = fmt.Sprintf(`select uin from experience_admin  where boardId = %d`, boardId)
	rows, err = inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		if uid == uin {
			hasPermission = true
		}
	}

	if uin == 100001 {
		hasPermission = true
	}

	return
}
