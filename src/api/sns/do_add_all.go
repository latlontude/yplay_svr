package sns

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type AddAllReq struct {
}

type AddAllRsp struct {
}

func doAddAll(req *AddAllReq, r *http.Request) (rsp *AddAllRsp, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}
	sql := fmt.Sprintf(`select uin from profiles`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	uids := make([]int64, 0)
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		AddCustomServiceAccount(uid)
		uids = append(uids, uid)
	}
	log.Debugf("uids = %v", uids)
	return
}
