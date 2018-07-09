package user

import (
	"api/geneqids"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/cache"
	"svr/st"
	"time"
)

type UpdateSchoolInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SchoolId   int    `schema:"schoolId"`
	Grade      int    `schema:"grade"`
	SchoolName string `schema:"schoolName"`

	DeptId   int    `schema:"deptId"`
	DeptName string `schema:"deptName"`
	EnrollmentYear  int `schema:"enrollmentYear"`

	Flag int `schema:"flag"` //1表示要记入修改次数限制中
}

type UpdateSchoolInfoRsp struct {
}

func doUpdateSchoolInfo(req *UpdateSchoolInfoReq, r *http.Request) (rsp *UpdateSchoolInfoRsp, err error) {

	log.Errorf("uin %d, UpdateSchoolInfoReq %+v", req.Uin, req)

	err = UpdateUserSchoolInfo(req.Uin, req.SchoolId, req.SchoolName, req.Grade, req.DeptId, req.DeptName, req.EnrollmentYear,req.Flag)
	if err != nil {
		log.Errorf("uin %d, UpdateSchoolInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	log.Errorf("uin %d, UpdateSchoolInfoRsp succ, %+v", req.Uin, rsp)

	return
}

//正常情况学校 学校 + 年级 + 院系 是一起修改不会只修改年级
func UpdateUserSchoolInfo(uin int64, schoolId int, schoolName string, grade int, deptId int, deptName string, enrollmentYear int ,flag int) (err error) {

	log.Errorf("start UpdateUserSchoolInfo uin:%d, schoolId:%d, schoolName:%s, grade:%d, deptId:%d, deptName:%s flag:%d", uin, schoolId, schoolName, grade, deptId, deptName, flag)

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "uin invalid")
		log.Errorf(err.Error())
		return
	}

	//grade 1~3 4初中或者高中毕业或者大四 100大学毕业
	if (grade != constant.ENUM_USER_GRADE_GRADUATE) && (grade > constant.ENUM_USER_GRADE_5 || grade < 0) {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "grade invalid")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	if _, ok := cache.SCHOOLS[schoolId]; !ok && !(schoolId <= 9999999 && schoolId >= 9999997) {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "schoolId invalid")
		log.Errorf(err.Error())
		return
	}

	if schoolId <= 9999999 && schoolId >= 9999997 { //999999[7~9] 代表用户自己输入学校 初中/高中/大学

		log.Errorf("uin:%d, pending schoolName:%s ", uin, schoolName)
		stmt, err1 := inst.Prepare(`insert into pendingSchool values(?, ?, ?, ?, ?, ?) on duplicate key update schoolId = ?, schoolName = ?, status = ?, ts = ?`)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_PREPARE, err1.Error())
			log.Error(err)
			return
		}
		defer stmt.Close()

		ts := time.Now().Unix()
		_, err1 = stmt.Exec(0, uin, schoolId, schoolName, 0, ts, schoolId, schoolName, 0, ts)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err1.Error())
			log.Error(err.Error())
			return
		}

	}
	log.Errorf("uin:%d, pending schoolName:%s ", uin, schoolName)
	var modQutoaInfo *st.ProfileModQuotaInfo

	//账号注册后的修改 要记入修改次数
	if flag > 0 {

		modQutoaInfo, err = st.GetUserProfileModQuotaInfo(uin, constant.ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		if modQutoaInfo.LeftCnt <= 0 {
			err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
			log.Errorf(err.Error())
			return
		}
	}

	var sql string

	if info, ok := cache.SCHOOLS[schoolId]; ok {

		//如果不是大学 或者大学的deptId为0
		//则不会更改学院信息
		if info.SchoolType != constant.ENUM_SCHOOL_TYPE_UNIVERSITY || deptId == 0 {

			//更新学校信息
			sql = fmt.Sprintf(`update profiles set grade = %d,schoolId = %d,schoolType = %d,schoolName = "%s",
							country = "%s",province = "%s",city = "%s",enrollmentYear = %d where uin = %d`,
				grade, info.SchoolId, info.SchoolType, info.SchoolName, info.Country, info.Province, info.City, enrollmentYear,uin)

			//不更新grade
			if grade == 0 {

				sql = fmt.Sprintf(`update profiles set schoolId = %d, schoolType = %d,schoolName = "%s",
							country = "%s",province = "%s",city = "%s" ,enrollmentYear = %d where uin = %d`,
					info.SchoolId, info.SchoolType, info.SchoolName, info.Country, info.Province, info.City,enrollmentYear, uin)
			}

		} else {
			//更新学校信息
			sql = fmt.Sprintf(`update profiles set grade = %d,schoolId = %d,schoolType = %d,schoolName = "%s",deptId = %d,
							deptName = "%s",country = "%s",province = "%s",city = "%s" ,enrollmentYear = %d where uin = %d`,
				grade, info.SchoolId, info.SchoolType, info.SchoolName, deptId, deptName, info.Country, info.Province, info.City, enrollmentYear,uin)

			//不更新grade
			if grade == 0 {
				sql = fmt.Sprintf(`update profiles set schoolId = %d,schoolType = %d,schoolName = "%s",deptId = %d,deptName = "%s",
							country = "%s",province = "%s",city = "%s" ,enrollmentYear = %d where uin = %d`,
					info.SchoolId, info.SchoolType, info.SchoolName, deptId, deptName, info.Country, info.Province, info.City,enrollmentYear, uin)
			}
		}
		log.Errorf("sql:%s",sql)

	} else {

		// 该学校待审核
		//ischoolId <= 9999999 && schoolId >= 999999

		tschoolType := 0

		if schoolId == 9999999 {
			tschoolType = 3
		} else if schoolId == 9999998 {
			tschoolType = 2
		} else if schoolId == 9999997 {
			tschoolType = 1
		}

		sql = fmt.Sprintf(`update profiles set grade = %d, schoolId = %d, schoolType = %d, schoolName = "%s" ,enrollmentYear = %d where uin = %d`,
			grade, schoolId, tschoolType, schoolName,enrollmentYear, uin)
		if grade == 0 {
			sql = fmt.Sprintf(`update profiles set schoolId = %d, schoolType = %d, schoolName = "%s" ,enrollmentYear = %d where uin = %d`,
				schoolId, tschoolType, schoolName,enrollmentYear, uin)
		}
		log.Errorf("sql:%s",sql)

	}


	log.Errorf("sql:%s",sql)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//账号注册后的修改 要记入修改次数
	if flag > 0 {
		modDesc := fmt.Sprintf("schoolId:%d, grade:%d", schoolId, grade)
		go st.AddProfileModRecordInfo(uin, constant.ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE, modDesc)

		//修改性别重新生成答题列表
		go geneqids.Gene(uin)
	}

	log.Errorf("end UpdateUserSchoolInfo")
	return
}
