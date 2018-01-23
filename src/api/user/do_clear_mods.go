package user

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type ClearModsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	UserName string `schema:"userName"`
}

type ClearModsRsp struct {
}

func doClearMods(req *ClearModsReq, r *http.Request) (rsp *ClearModsRsp, err error) {

	log.Errorf("uin %d, ClearModsReq %+v", req.Uin, req)

	err = ClearMods(req.UserName)
	if err != nil {
		log.Errorf("uin %d, ClearModsRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &ClearModsRsp{}

	log.Errorf("uin %d, ClearModsRsp succ, %+v", req.Uin, rsp)

	return
}

func ClearMods(userName string) (err error) {

	if len(userName) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin from profiles where userName = "%s"`, userName)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false
	var uin int64

	for rows.Next() {
		rows.Scan(&uin)
		find = true
	}

	if !find || uin == 0 {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "permission denied")
		log.Error(err.Error())
		return
	}

	//注销账号
	sqls := make([]string, 0)
	sqls = append(sqls, fmt.Sprintf(`delete from profileModRecords where uin = %d`, uin))

	err = mydb.Exec(inst, sqls)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
