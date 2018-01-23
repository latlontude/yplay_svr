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
)

type UpdateSchoolInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SchoolId int `schema:"schoolId"`
	Grade    int `schema:"grade"`

	Flag int `schema:"flag"` //1表示要记入修改次数限制中
}

type UpdateSchoolInfoRsp struct {
}

func doUpdateSchoolInfo(req *UpdateSchoolInfoReq, r *http.Request) (rsp *UpdateSchoolInfoRsp, err error) {

	log.Errorf("uin %d, UpdateSchoolInfoReq %+v", req.Uin, req)

	err = UpdateUserSchoolInfo(req.Uin, req.SchoolId, req.Grade, req.Flag)
	if err != nil {
		log.Errorf("uin %d, UpdateSchoolInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	log.Errorf("uin %d, UpdateSchoolInfoRsp succ, %+v", req.Uin, rsp)

	return
}

func UpdateUserSchoolInfo(uin int64, schoolId int, grade int, flag int) (err error) {

	if uin == 0 {
		return
	}

	//grade 1~3 4初中或者高中毕业或者大四 100大学毕业
	if (grade != constant.ENUM_USER_GRADE_GRADUATE) && (grade > constant.ENUM_USER_GRADE_5 || grade < 0) {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "grade invalid")
		log.Errorf(err.Error())
		return
	}

	if _, ok := cache.SCHOOLS[schoolId]; !ok {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "schoolId not found")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

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

	info := cache.SCHOOLS[schoolId]

	//更新学校信息
	sql := fmt.Sprintf(`update profiles set 
							grade = %d, 
							schoolId = %d, 
							schoolType = %d, 
							schoolName = "%s", 
							country = "%s", 
							province = "%s", 
							city = "%s" 
							where uin = %d`,
		grade, info.SchoolId, info.SchoolType, info.SchoolName, info.Country, info.Province, info.City, uin)

	//不更新grade
	if grade == 0 {

		sql = fmt.Sprintf(`update profiles set 							
							schoolId = %d, 
							schoolType = %d, 
							schoolName = "%s", 
							country = "%s", 
							province = "%s", 
							city = "%s" 
							where uin = %d`,
			info.SchoolId, info.SchoolType, info.SchoolName, info.Country, info.Province, info.City, uin)

	}

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

	return
}
