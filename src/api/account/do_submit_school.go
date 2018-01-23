package account

import (
	// "common/constant"
	// "common/env"
	// "common/rest"
	// "common/token"
	"net/http"
)

type SubmitSchoolReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SchoolName string `schema:"schoolName"`
}

type SubmitSchoolRsp struct {
}

func doSubmitSchool(req *SubmitSchoolReq, r *http.Request) (rsp *SubmitSchoolRsp, err error) {

	log.Debugf("SubmitSchoolReq %+v", req)

	log.Debugf("SubmitSchoolRsp %+v", rsp)

	return
}
