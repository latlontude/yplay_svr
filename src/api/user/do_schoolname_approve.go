package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/cache"
	"svr/st"
)

type SchoolNameApproveReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type SchoolNameApproveRsp struct {
}

func doApproveSchoolName(req *SchoolNameApproveReq, r *http.Request) (rsp *SchoolNameApproveRsp, err error) {

	log.Errorf("start doApproveSchoolName uin:%d", req.Uin)

	info, err := ApproveSchoolName(req.Uin)
	if err != nil {
		log.Errorf("uin %d, SubmitApproveRsp error, %s", req.Uin, err.Error())
		return
	}
	rsp = &info
	log.Errorf("start doApproveSchoolName")
	return

}

func ApproveSchoolName(uin int64) (ret SchoolNameApproveRsp, err error) {

	log.Errorf("start ApproveSchoolName")

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select schoolId from pendingSchool where uin = %d and status = 1`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	var schoolId int
	for rows.Next() {
		rows.Scan(&schoolId)
	}

	if schoolId != 0 {

		sql := fmt.Sprintf(`select schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts from schools where schoolId = %d`, schoolId)
		rows, err1 := inst.Query(sql)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
			log.Errorf(err.Error())
			return
		}
		defer rows.Close()

		var schoolInfo st.SchoolInfo
		for rows.Next() {
			rows.Scan(&schoolInfo.SchoolId, &schoolInfo.SchoolType, &schoolInfo.SchoolName, &schoolInfo.Country, &schoolInfo.Province, &schoolInfo.City,
				&schoolInfo.Latitude, &schoolInfo.Longitude, &schoolInfo.Status, &schoolInfo.Ts)
		}
		cache.AddCacheSchool(schoolInfo)
	}

	log.Errorf("end ApproveSchoolName")
	return
}
