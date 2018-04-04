package activity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type BeSingerFansFromWxReq struct {
	OpenId   string `schema:"openId"`
	SingerId int    `schema:"singerId"`
}

type BeSingerFansFromWxRsp struct {
	Status int `json:"status"` // 1 成功成为歌手粉丝，2 已经是该歌手的粉丝，3 已经是其他歌手的粉丝
}

func doBeSingerFansFromWx(req *BeSingerFansFromWxReq, r *http.Request) (rsp *BeSingerFansFromWxRsp, err error) {
	log.Debugf("start doBeSingerFansFromWx openId:%s singerId:%d", req.OpenId, req.SingerId)
	status, err := BeSingerFansFromWx(req.OpenId, req.SingerId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &BeSingerFansFromWxRsp{status}
	log.Debugf("end doBeSingerFansFromWx status:%d", status)
	return
}

func BeSingerFansFromWx(openId string, singerId int) (status int, err error) {
	log.Debugf("start BeSingerFansFromWx openId:%s", openId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId, openId from ddsingerFansFromWx where status = 0 and openId = "%s" and singerId = %d`, openId, singerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		status = 2
		log.Debugf("openId:%s has been singer:%d fans", openId, singerId)
		return
	}

	sql = fmt.Sprintf(`select openId from ddsingerFansFromWx where status = 0 and openId = "%s"`, openId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		status = 3
		log.Debugf("openId:%s has been other singer fans", openId)
		return
	}

	stmt, err := inst.Prepare(`insert into ddsingerFansFromWx values(?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()
	stat := 0
	_, err = stmt.Exec(openId, singerId, stat, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	status = 1
	log.Debugf("end BeSingerFansFromWx")
	return
}
