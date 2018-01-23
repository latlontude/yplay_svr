package st

import (
	"common/constant"
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
