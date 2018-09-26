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
type UpdateAnOrderReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SchoolId  int `schema:"schoolId"`
	OrderId   int `schema:"orderId"`
	OrderType int `schema:"orderType"`
}

//返回分好的列表
type UpdateAnOrderRsp struct {
	Code int `json:"code"`
}

func doUpdateAnOrder(req *UpdateAnOrderReq, r *http.Request) (rsp *UpdateAnOrderRsp, err error) {

	log.Debugf("UpdateAnOrderReq : %+v", req)
	err = UpdateAnOrder(req.Uin, req.OrderType, req.OrderId, req.SchoolId)

	if err != nil {
		log.Debugf("dispatch err uin:%d, schoolId:%d orderId:%d", req.Uin, req.SchoolId, req.OrderId)
		rsp = &UpdateAnOrderRsp{-1}
		return
	}

	code := 0
	rsp = &UpdateAnOrderRsp{code}
	log.Debugf("DispatchOrderRsp:%+v", rsp)

	return
}

func UpdateAnOrder(uin int64, orderType int, orderId int, schoolId int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	now := time.Now().Unix()

	var sql string
	//跑腿者接单
	if orderType == 3 {
		sql = fmt.Sprintf(`update express_orderInfo set status = %d,sendTs = %d where orderId = %d and schoolId = %d`,
			orderType, now, orderId, schoolId)
	} else if orderType == 4 {
		sql = fmt.Sprintf(`update express_orderInfo set status = %d,finishTs = %d where orderId = %d and schoolId = %d`,
			orderType, now, orderId, schoolId)
	}

	log.Debugf("updateAnOrder_sql:%s", sql)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}
