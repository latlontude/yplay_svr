package account

import (
	"net/http"
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

	rsp = &GetDeptsBySchoolRsp{depts}

	log.Debugf("uin %d, GetDeptsBySchoolRsp succ, %+v", req.Uin, rsp)

	return
}
