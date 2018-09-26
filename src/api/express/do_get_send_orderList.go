package express

import (
	"net/http"
)

type GetSendOrderListReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	OrderType int `schema:"orderType"`
	SchoolId  int `schema:"schoolId"`
	PageNum   int `schema:"pageNum"`
	PageSize  int `schema:"pageSize"`
}

type GetSendOrderListRsp struct {
	OrderList []*OrderInfo `json:"orderList"`
}

func doGetSendOrderList(req *GetSendOrderListReq, r *http.Request) (rsp *GetSendOrderListRsp, err error) {

	log.Debugf("GetSendOrderListReq req:%+v", req)
	orderList, err := GetSendOrderList(req.Uin, req.SchoolId, req.OrderType)

	if err != nil {
		log.Debugf("GetSendOrderList error ,err:%+v", err)
		return
	}

	rsp = &GetSendOrderListRsp{orderList}

	log.Debugf("GetSendOrderListReq , rsp:%+v", rsp)

	return
}
