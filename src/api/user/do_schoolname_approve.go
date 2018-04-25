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
	User     int64 `schema:"user"`
	SchoolId int   `schema:"schoolId"`
}

type SchoolNameApproveRsp struct {
}

func doApproveSchoolName(req *SchoolNameApproveReq, r *http.Request) (rsp *SchoolNameApproveRsp, err error) {

	log.Errorf("doApproveSchoolName req %+v", req)

	err = ApproveSchoolName(req.User, req.SchoolId)
	if err != nil {
		log.Errorf("SubmitApproveSchoolRsp error %s", err.Error())
		return
	}

	rsp = &SchoolNameApproveRsp{}

	log.Errorf("SubmitApproveSchoolRsp succ")
	return

}

func ApproveSchoolName(user int64, schoolId int) (err error) {

	if user == 0 || schoolId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid user or schoolId")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	info, err := st.GetSchoolInfo(schoolId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//学校不存在
	if info.SchoolId == 0 {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "schoolId zero!")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select id from pendingSchool where uin = %d  and schoolId >= 9999997 and schoolId <= 9999999`, user)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		var t int
		rows.Scan(&t)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "record not found in pendingschool")
		log.Errorf(err.Error())
		return
	}

	//查询用户的资料是否已经设置了学校
	ui, err := st.GetUserProfileInfo(user)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	sqls := make([]string, 0)

	if ui.SchoolId >= 9999997 && ui.SchoolId <= 9999999 {
		//学校处于审核中

		sql1 := fmt.Sprintf(`update profiles set schoolId = %d, schoolType = %d, schoolName = "%s", country = "%s", province = "%s", city = "%s", deptId = 0, deptName = "" where uin = %d`,
			info.SchoolId, info.SchoolType, info.SchoolName, info.Country, info.Province, info.City, user)

		sql2 := fmt.Sprintf(`update pendingSchool set status = 1 where uin = %d`, user)

		sqls = append(sqls, sql1)
		sqls = append(sqls, sql2)

	} else if ui.SchoolId > 0 {
		//学校已经修改过了

		sql2 := fmt.Sprintf(`update pendingSchool set status = 2 where uin = %d`, user)

		sqls = append(sqls, sql2)

	} else {
		// schoolId == 0
	}

	if len(sqls) > 0 {
		err = mydb.Exec(inst, sqls)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
	}

	if _, ok := cache.SCHOOLS[schoolId]; !ok {
		cache.AddCacheSchool(*info)
	}

	return
}
