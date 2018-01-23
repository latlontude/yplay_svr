package submit

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type SubmitDeleteReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SubmitId int    `schema:"submitId"`
}

type SubmitDeleteRsp struct {
}

func doSubmitDelete(req *SubmitDeleteReq, r *http.Request) (rsp *SubmitDeleteRsp, err error) {

	log.Errorf("uin %d, SubmitDeleteReq %+v", req.Uin, req)

	err = SubmitDelete(req.Uin, req.SubmitId)
	if err != nil {
		log.Errorf("uin %d, SubmitDeleteRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitDeleteRsp{}

	log.Errorf("uin %d, SubmitDeleteRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitDelete(uin int64, submitId int) (err error) {

	if uin == 0 || submitId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	sql := fmt.Sprintf(`select qtext from submitQuestions where id = %d and uin = %d and status = %d`, submitId, uin, 2)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		var tmp string
		rows.Scan(&tmp)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "res not found")
		log.Errorf(err.Error())
		return
	}

	//更新审核题目信息和状态
	sql = fmt.Sprintf(`delete from submitQuestions where id = %d and uin = %d and status = %d`, submitId, uin, 2)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
