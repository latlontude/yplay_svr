package express

import (
	"net/http"
)

type GetBoardOrderListReq struct {
	Uin       int64  `schema:"uin"`
	Token     string `schema:"token"`
	Ver       int    `schema:"ver"`
	SchoolId  int    `schema:"schoolId"`
	OrderType int    `schema:"orderType"`
}

type GetBoardOrderListRsp struct {
	OrderList []*OrderInfo `json:"orderList"`
}

func doGetBoardOrderList(req *GetBoardOrderListReq, r *http.Request) (rsp *GetBoardOrderListRsp, err error) {

	log.Debugf("GetBoardOrderListReq req:%+v", req)
	orderList, err := GetBoardOrderList(req.Uin, req.SchoolId, req.OrderType)

	if err != nil {
		log.Debugf("GetBoardOrderList error ,err:%+v", err)
		return
	}

	rsp = &GetBoardOrderListRsp{orderList}

	log.Debugf("GetBoardOrderListRsp , rsp:%+v", rsp)

	return
}
