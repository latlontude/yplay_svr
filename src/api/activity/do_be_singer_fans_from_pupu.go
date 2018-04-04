package activity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type BeSingerFansFromPupuReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SingerId int    `schema:"singerId"`
}

type BeSingerFansFromPupuRsp struct {
	Status int `json:"status"` // 1 成功成为歌手粉丝，2 已经是该歌手的粉丝，3 已经是其他歌手的粉丝
}

func doBeSingerFansFromPupu(req *BeSingerFansFromPupuReq, r *http.Request) (rsp *BeSingerFansFromPupuRsp, err error) {
	log.Debugf("start doBeSingerFansFromPupu uin:%d singerId:%d", req.Uin, req.SingerId)
	status, err := BeSingerFansFromPupu(req.Uin, req.SingerId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &BeSingerFansFromPupuRsp{status}
	log.Debugf("end doBeSingerFansFromPupu status:%d", status)
	return
}

func BeSingerFansFromPupu(uin int64, singerId int) (status int, err error) {
	log.Debugf("start BeSingerFansFromPupu uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId, uin from ddsingerFansFromPupu where status = 0 and uin = %d and singerId = %d`, uin, singerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		status = 2
		log.Debugf("uin:%d has been singer:%d fans", uin, singerId)
		return
	}

	sql = fmt.Sprintf(`select uin from ddsingerFansFromPupu where status = 0 and uin = %d`, uin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		status = 3
		log.Debugf("uin:%d has been other singer fans", uin)
		return
	}

	stmt, err := inst.Prepare(`insert into ddsingerFansFromPupu values(?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()
	stat := 0
	_, err = stmt.Exec(uin, singerId, stat, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	status = 1
	log.Debugf("end BeSingerFansFromPupu")
	return
}
