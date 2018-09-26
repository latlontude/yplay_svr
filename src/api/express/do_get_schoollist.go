package express

import (
	"net/http"
)

type GetSchoolListReq struct {
	Uin int64 `schema:"uin"`
}

type SchoolInfo struct {
	SchoolId   int    `json:"schoolId"`
	SchoolName string `json:"schoolName"`
}
type GetSchoolListRsp struct {
	SchoolList []*SchoolInfo `json:"schoolList"`
}

func doGetSchoolList(req *GetSchoolListReq, r *http.Request) (rsp *GetSchoolListRsp, err error) {
	log.Debugf("GetSchoolListReq:%+v", req)
	schoolList, err := GetSchoolList()
	if err != nil {
		log.Errorf("GetSchoolList error,err:%+v", err)
	}
	rsp = &GetSchoolListRsp{schoolList}
	log.Debugf("GetPositionListRsp : %+v", rsp)
	return
}

//获取某个学校的位置信息
func GetSchoolList() (schoolList []*SchoolInfo, err error) {
	var schoolInfo SchoolInfo
	schoolInfo.SchoolId = 78629
	schoolInfo.SchoolName = "中国地质大学"
	schoolList = make([]*SchoolInfo, 0)
	schoolList = append(schoolList, &schoolInfo)

	var schoolInfo2 SchoolInfo
	schoolInfo2.SchoolId = 77557
	schoolInfo2.SchoolName = "黑龙江大学"
	schoolList = append(schoolList, &schoolInfo2)

	var schoolInfo3 SchoolInfo
	schoolInfo3.SchoolId = 80464
	schoolInfo3.SchoolName = "青海兽医"
	schoolList = append(schoolList, &schoolInfo3)
	return
}
