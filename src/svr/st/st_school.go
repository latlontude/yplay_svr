package st

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

type SchoolInfo struct {
	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"school"`

	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Status int `json:"status"`
	Ts     int `json:"ts"`
}

type SchoolInfo2 struct {
	SchoolId   int     `json:"schoolId"`
	SchoolType int     `json:"schoolType"`
	SchoolName string  `json:"school"`
	Country    string  `json:"country"`
	Province   string  `json:"province"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Status     int     `json:"status"`
	Ts         int     `json:"ts"`
	MemberCnt  int     `json:"memberCnt"`
}

type DeptInfo struct {
	DeptId   int    `json:"deptId"`
	DeptName string `json:"deptName"`
}

func (this *DeptInfo) String() string {

	return fmt.Sprintf(`DeptInfo{DeptId:%d DeptName:%s}`, this.DeptId, this.DeptName)
}

func GetGradeDescBySchool(schoolType int, grade int) (desc string) {

	if schoolType == constant.ENUM_SCHOOL_TYPE_UNIVERSITY {

		if grade == 1 {
			desc = "2017"
		} else if grade == 2 {
			desc = "2016"
		} else if grade == 3 {
			desc = "2015"
		} else if grade == 4 {
			desc = "2014"
		} else if grade == 5 {
			desc = "2013"
		} else if grade == 100 {
			desc = "2018"
		}

	} else if schoolType == constant.ENUM_SHOOL_TYPE_HIGH {

		if grade == 1 {
			desc = "2017"
		} else if grade == 2 {
			desc = "2016"
		} else if grade == 3 {
			desc = "2015"
		} else if grade == 100 {
			desc = "2018"
		}

	} else if schoolType == constant.ENUM_SHOOL_TYPE_JUNIOR {

		if grade == 1 {
			desc = "2017"
		} else if grade == 2 {
			desc = "2016"
		} else if grade == 3 {
			desc = "2015"
		} else if grade == 100 {
			desc = "2018"
		}
	}

	return
}

func GetDeptsBySchool(schoolId int) (infos []*DeptInfo, err error) {

	infos = make([]*DeptInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}
	sql := fmt.Sprintf(`select deptId, deptName from departments where schoolId = %d order by convert(deptName using gbk)asc`, schoolId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	var otherDeptInfo DeptInfo
	for rows.Next() {

		var info DeptInfo

		rows.Scan(
			&info.DeptId,
			&info.DeptName)

		if info.DeptName == "其它院系" {
			otherDeptInfo = info
		} else {
			infos = append(infos, &info)
		}
	}

	infos = append(infos, &otherDeptInfo)
	return
}

func GetSchoolInfo(schoolId int) (info *SchoolInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts from schools where schoolId = %d`, schoolId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	info = &SchoolInfo{}

	find := false
	for rows.Next() {
		rows.Scan(&info.SchoolId, &info.SchoolType, &info.SchoolName, &info.Country, &info.Province, &info.City, &info.Latitude, &info.Longitude, &info.Status, &info.Ts)
		find = true
		break
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, fmt.Sprintf("schoolId %d not found", schoolId))
		log.Errorf(err.Error())
		return
	}

	return
}
