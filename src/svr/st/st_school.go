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

func GetGradeDescBySchool(schoolType int, grade int) (desc string) {

	if schoolType == constant.ENUM_SCHOOL_TYPE_UNIVERSITY {

		if grade == 1 {
			desc = "大一"
		} else if grade == 2 {
			desc = "大二"
		} else if grade == 3 {
			desc = "大三"
		} else if grade == 4 {
			desc = "大四"
		} else if grade == 5 {
			desc = "大五"
		} else if grade == 100 {
			desc = "大学毕业"
		}

	} else if schoolType == constant.ENUM_SHOOL_TYPE_HIGH {

		if grade == 1 {
			desc = "高一"
		} else if grade == 2 {
			desc = "高二"
		} else if grade == 3 {
			desc = "高三"
		} else if grade == 100 {
			desc = "高中毕业"
		}

	} else if schoolType == constant.ENUM_SHOOL_TYPE_JUNIOR {

		if grade == 1 {
			desc = "初一"
		} else if grade == 2 {
			desc = "初二"
		} else if grade == 3 {
			desc = "初三"
		} else if grade == 100 {
			desc = "初中毕业"
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
