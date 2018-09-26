//墙主查看 该学校的跑腿者

package express

import (
	"net/http"
)

type AddSenderInfoReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SchoolId int    `schema:"schoolId"`
	UserName string `schema:"userName"`
	Name     string `schema:"name"`
	Phone    string `schema:"phone"`
}

type AddSenderInfoRsp struct {
	SenderInfo SenderInfo `json:"senderInfo"`
}

func doAddSenderInfo(req *AddSenderInfoReq, r *http.Request) (rsp *AddSenderInfoRsp, err error) {

	log.Debugf("AddSenderInfoReq req:%+v", req)
	senderInfo, err := AddSenderInfo(req.Uin, req.SchoolId, req.UserName, req.Name, req.Phone)

	if err != nil {
		log.Debugf("AddSenderInfo error ,err:%+v", err)
		return
	}

	rsp = &AddSenderInfoRsp{senderInfo}

	log.Debugf("AddSenderInfoRsp , rsp:%+v", rsp)

	return
}
