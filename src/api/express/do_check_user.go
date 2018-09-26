package express

import (
	"net/http"
	"svr/st"
)

//把一车订单派给某个人
type CheckUserReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SchoolId int    `schema:"schoolId"`
}

type CheckUserRsp struct {
	IsBoardOwner bool `json:"isBoardOwner"`
	IsSender     bool `json:"isSender"`
}

func doCheckUser(req *CheckUserReq, r *http.Request) (rsp *CheckUserRsp, err error) {

	log.Debugf("CheckUserReq:%+v", req)
	isBoardOwner, isSender, err := st.CheckExpressUser(req.Uin, req.SchoolId)
	if err != nil {
		log.Debugf("CheckUser err ")
		return
	}
	rsp = &CheckUserRsp{isBoardOwner, isSender}

	log.Debugf("CheckUserRsp:%+v", rsp)
	return
}
