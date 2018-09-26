package express

import (
	"net/http"
	"time"
)

type GetFeeReq struct {
	SchoolId    int    `schema:"schoolId"`
	SendAddr    string `schema:"sendAddr"`
	ReceiveAddr string `schema:"receiveAddr"`
	ParcelSize  int    `schema:"parcelSize"`
}

type GetFeeRsp struct {
	Fee int `json:"fee"`
}

func doGetFee(req *GetFeeReq, r *http.Request) (rsp *GetFeeRsp, err error) {
	log.Debugf("GetFeeReq:%+v", req)
	fee, err := GetFee(req.SchoolId, req.SendAddr, req.ReceiveAddr, req.ParcelSize)
	if err != nil {
		log.Errorf("GetFee error,err:%+v", err)
	}
	rsp = &GetFeeRsp{fee}
	log.Debugf("GetFeeRsp : %+v", rsp)
	return
}

//获取某个学校的位置信息
func GetFee(schoolId int, sendAddr string, receiveAddr string, parcelSize int) (fee int, err error) {
	fee = 500
	return
}

func GetArrivalTime(arrival int) (arrivalTs int64) {
	now := time.Now()
	currentTimeData := time.Date(now.Year(), now.Month(), now.Day(), arrival, 0, 0, 0, time.Local) //获取当前时间，返回当前时间Time
	arrivalTs = currentTimeData.Unix()                                                             //转化为时间戳 类型是int64

	log.Debugf("arrival:%d", arrival)

	return
}
