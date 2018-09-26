//墙主查看 该学校的跑腿者

package express

import (
	"net/http"
)

type GetSenderListReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SchoolId int    `schema:"schoolId"`
}

type GetSenderListRsp struct {
	SenderList []*SenderInfo `json:"senderList"`
}

func doGetSenderList(req *GetSenderListReq, r *http.Request) (rsp *GetSenderListRsp, err error) {

	log.Debugf("GetBoardSenderListReq req:%+v", req)
	senderList, err := GetSenderList(req.Uin, req.SchoolId)

	if err != nil {
		log.Debugf("GetBoardSenderList error ,err:%+v", err)
		return
	}

	rsp = &GetSenderListRsp{senderList}

	log.Debugf("GetBoardSenderListRsp , rsp:%+v", rsp)

	return
}
