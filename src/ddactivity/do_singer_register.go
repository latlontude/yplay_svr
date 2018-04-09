package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type SingerRegisterReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type SingerRegisterRsp struct {
}

func doSingerRegister(req *SingerRegisterReq, r *http.Request) (rsp *SingerRegisterRsp, err error) {
	log.Debugf("start doRegisterSinger uin:%d", req.Uin)
	err = SingerRegister(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &SingerRegisterRsp{}
	log.Debugf("end doRegisterSinger ")
	return
}

func SingerRegister(uin int64) (err error) {
	log.Debugf("start SingerRegister uin:%d", uin)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//看看是否在配置的学校列表中
	log.Debugf("openSchools:%+v", OpenSchools)
	log.Debugf("ui schoolId:%d", ui.SchoolId)

	find := false
	for sid, _ := range OpenSchools {
		if ui.SchoolId == sid {
			find = true
			break
		}
	}

	//没有参加活动的权限
	if !find {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Debugf("user:%d school does not fit", uin)
		return
	}

	if ui.Uin == 0 || len(ui.NickName) == 0 || len(ui.DeptName) == 0 || len(ui.HeadImgUrl) == 0 || ui.Grade == 0 || ui.Gender == 0 {
		log.Debugf("ui:%+v register info is incomplete")
		return
	}

	sql := fmt.Sprintf(`select uin from ddsingers where uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	hasRegistered := false
	for rows.Next() {
		hasRegistered = true
		break
	}
	if hasRegistered {
		log.Debugf("uin:%d has register", uin)
		return
	}

	stmt, err := inst.Prepare(`insert into ddsingers values(?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()
	status := 0
	_, err = stmt.Exec(0, ui.Uin, ui.NickName, ui.HeadImgUrl, ui.Gender, ui.DeptName, ui.Grade, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end SingerRegister")
	return
}
