package account

import (
	"net/http"
	"svr/cache"
	"svr/st"
)

type GetDeptsBySchoolReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SchoolId int `schema:"schoolId"`
}

type GetDeptsBySchoolRsp struct {
	Depts []*st.DeptInfo `json:"depts"`
}

func doGetDeptsBySchool(req *GetDeptsBySchoolReq, r *http.Request) (rsp *GetDeptsBySchoolRsp, err error) {

	log.Debugf("uin %d, GetDeptsBySchoolReq %+v", req.Uin, req)

	depts, err := st.GetDeptsBySchool(req.SchoolId)
	if err != nil {
		log.Errorf("uin %d, GetDeptsBySchoolRsp error, %s", req.Uin, err.Error())
		return
	}

	//如果发现某个学校院系是空，则补充一个默认的
	if len(depts) == 0 {
		if _, ok := cache.SCHOOLS[req.SchoolId]; ok {
			deptId := req.SchoolId*1000 + 1
			depts = make([]*st.DeptInfo, 0)
			depts = append(depts, &st.DeptInfo{deptId, "其他院系"})
		}
	}

	rsp = &GetDeptsBySchoolRsp{depts}

	log.Debugf("uin %d, GetDeptsBySchoolRsp succ, %+v", req.Uin, rsp)

	return
}
