//墙主查看 该学校的跑腿者

package express

import (
	"net/http"
)

type GetOrderIdReq struct {
	SchoolId int    `schema:"schoolId"`
	Openid   string `schema:"openid"`
}

type GetOrderIdRsp struct {
	OrderId int64 `json:"orderId"`
}

func doGetOrderId(req *GetOrderIdReq, r *http.Request) (rsp *GetOrderIdRsp, err error) {

	log.Debugf("GetOrderIdReq req:%+v", req)
	orderId, err := GetOrderId(req.SchoolId, req.Openid)

	if err != nil {
		log.Debugf("GetOrderIdReq error ,err:%+v", err)
		return
	}

	rsp = &GetOrderIdRsp{orderId}

	log.Debugf("GetBoardSenderListRsp , rsp:%+v", rsp)

	return
}
