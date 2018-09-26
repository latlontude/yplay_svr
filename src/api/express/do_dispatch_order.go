package express

//墙主派单

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

//把一车订单派给某个人
type DispatchOrderReq struct {
	Uin       int64  `schema:"uin"`
	Token     string `schema:"token"`
	Ver       int    `schema:"ver"`
	SchoolId  int    `schema:"schoolId"`
	OrderIds  string `schema:"orderIds"`
	OrderType int    `schema:"orderType"`
	ToUin     int64  `schema:"toUin"`
}

//返回分好的列表
type DispatchOrderRsp struct {
	Code int `json:"code"`
}

func doDispatchOrder(req *DispatchOrderReq, r *http.Request) (rsp *DispatchOrderRsp, err error) {

	err = DispatchOrder(req.Uin, req.ToUin, req.OrderType, req.OrderIds, req.SchoolId)
	if err != nil {
		log.Debugf("dispatch err uin:%d,toUin:%d , schoolId:%d orderIds:%s", req.Uin, req.ToUin, req.SchoolId, req.OrderIds)
		rsp = &DispatchOrderRsp{-1}
		return
	}
	code := 0
	rsp = &DispatchOrderRsp{code}
	log.Debugf("DispatchOrderRsp:%+v", rsp)

	return
}

func DispatchOrder(uin int64, toUin int64, orderType int, orderIds string, schoolId int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	now := time.Now().Unix()
	sql := fmt.Sprintf(`update express_orderInfo 
set status = %d,senderUid = %d ,dispatchTs = %d where orderId in (%s) and schoolId = %d`,
		orderType, toUin, now, orderIds, schoolId)

	log.Debugf("sql:%s", sql)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
