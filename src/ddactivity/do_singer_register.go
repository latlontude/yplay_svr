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

	PersonalName           string `schema:"personalName"`
	ActiveHeadImgUrl       string `schema:"activeHeadImgUrl"`
	RankActiveHeadImgUrl   string `schema:"rankActiveHeadImgUrl"`
	SingerDetailInfoImgUrl string `schema:"singerDetailInfoImgUrl"`
	DeptName               string `schema:"deptName"`
	Declaration            string `schema:"declaration"`
}

type Singer struct {
	Uin                    int64  `json:"uin"`
	PersonalName           string `json:"personalName"`
	ActiveHeadImgUrl       string `json:"activeHeadImgUrl"`
	RankActiveHeadImgUrl   string `json:"rankActiveHeadImgUrl"`
	SingerDetailInfoImgUrl string `json:"singerDetailInfoImgUrl"`
	DeptName               string `json:"deptName"`
	Declaration            string `json:"declaration"`
}

type SingerRegisterRsp struct {
}

func doSingerRegister(req *SingerRegisterReq, r *http.Request) (rsp *SingerRegisterRsp, err error) {
	log.Debugf("start doRegisterSinger uin:%d", req.Uin)

	err = SingerRegister(Singer{req.Uin, req.PersonalName, req.ActiveHeadImgUrl, req.RankActiveHeadImgUrl, req.SingerDetailInfoImgUrl, req.DeptName, req.Declaration})
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &SingerRegisterRsp{}
	log.Debugf("end doRegisterSinger ")
	return
}

func SingerRegister(singer Singer) (err error) {
	log.Debugf("start SingerRegister uin:%d,")
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	ui, err := st.GetUserProfileInfo(singer.Uin)
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
		log.Debugf("user:%d school does not fit", singer.Uin)
		return
	}

	sql := fmt.Sprintf(`select uin from ddsingers where uin = %d`, singer.Uin)
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
		log.Debugf("uin:%d has register", singer.Uin)
		return
	}

	stmt, err := inst.Prepare(`insert into ddsingers values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()
	status := 0
	_, err = stmt.Exec(0, singer.Uin, singer.PersonalName, singer.ActiveHeadImgUrl, singer.RankActiveHeadImgUrl, singer.SingerDetailInfoImgUrl, singer.DeptName, singer.Declaration, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end SingerRegister")
	return
}
