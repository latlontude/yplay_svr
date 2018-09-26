package express

import (
	"net/http"
)

type OrderInfo struct {
	OrderId     string      `json:"orderId"`
	Openid      string      `json:"openid"`
	SchoolId    int         `json:"schoolId"`
	Name        string      `json:"name"`
	Phone       string      `json:"phone"`
	ParcelInfo  string      `json:"parcelInfo"`
	ParcelSize  int         `json:"parcelSize"`
	SendAddr    string      `json:"sendAddr"`
	ReceiveAddr string      `json:"receiveAddr"`
	OrderTs     int64       `json:"orderTs"`
	ArrivalTs   int64       `json:"arrivalTs"`
	DispatchTs  int64       `json:"dispatchTs"`
	SendTs      int64       `json:"sendTs"`
	FinishTs    int64       `json:"finishTs"`
	Status      int         `json:"status"`
	Fee         int         `json:"fee"`
	SenderInfo  *SenderInfo `json:"senderInfo"`
}

type GetOrderListReq struct {
	SchoolId int    `schema:"schoolId"`
	Openid   string `schema:"openid"`
}

type GetOrderListRsp struct {
	OrderList []*OrderInfo `json:"orderList"`
}

func doGetMyOrderList(req *GetOrderListReq, r *http.Request) (rsp *GetOrderListRsp, err error) {

	log.Debugf("GetMyOrderListReq req:%+v", req)
	orderList, err := GetMyOrderList(req.Openid, req.SchoolId)

	if err != nil {
		log.Debugf("GetMyOrderListByStatus error ,err:%+v", err)
		return
	}

	rsp = &GetOrderListRsp{orderList}

	log.Debugf("GetMyOrderListRsp , rsp:%+v", rsp)

	return
}
