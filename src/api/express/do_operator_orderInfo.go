package express

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"time"
)

func GetOrderId(schoolId int, openid string) (orderId int64, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into express_orderInfo(
				orderId,
				prepay_id,
				openId,
				schoolId,
				name,
				phone,
				parcelSize,
				parcelInfo,
				sendAddr,
				receiveAddr,
				orderTs,
				arrivalTs,
				dispatchTs,
				sendTs,
				finishTs ,
				status,
				fee,
				senderUid) 
		values(?, ?, ?, ?, ?, ?, ?, ? ,?, ?, ?, ?, ?, ?, ?, ? ,? ,?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(0, 0, openid, schoolId, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	orderId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	return
}
func InsertOrderInfo(req *OrderReq, prepayId string, fee int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	now := time.Now().Unix()

	var sql string

	arrivalTs := GetArrivalTime(req.ArrivalTs)
	sql = fmt.Sprintf(`update express_orderInfo set 
									prepay_id = '%s',
									name = '%s',
									phone = '%s',
									parcelSize = %d,
									parcelInfo = '%s',
									sendAddr = '%s',
									receiveAddr = '%s',
									orderTs = %d,
									arrivalTs = %d,
									dispatchTs = %d,
									sendTs = %d,
									finishTs =%d,
									status = %d,
									fee = %d,
									senderUid = %d
									where orderId = %d and schoolId = %d`,
		prepayId, req.Name, req.Phone, req.ParcelSize, req.ParcelInfo, req.SendAddr, req.ReceiveAddr, now,
		arrivalTs, 0, 0, 0, 0, fee, 0, req.OrderId, req.SchoolId)

	log.Debugf("update express_orderInfo:%s", sql)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

func GetOrderListByOrderType(openid string, orderType int) (orderList []*OrderInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	var sql, statusString string
	if orderType == 1 {
		statusString = fmt.Sprintf("(status != 1 and status != 4)") //下单和被接单
	} else {
		statusString = fmt.Sprintf("(status = 1 or status = 4)") //已完成和已取消的订单 都放到已结束中
	}

	sql = fmt.Sprintf(`select orderId,openId,schoolId,name,phone,parcelSize,parcelInfo,sendAddr,
recieveAddr,orderTs,arrivalTs,dispatchTs,sendTs,finishTs ,status,fee from express_orderInfo 
where openid = '%s' and %s`, openid, statusString)

	log.Debugf("orderSql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	orderList = make([]*OrderInfo, 0)

	for rows.Next() {
		var orderInfo OrderInfo
		rows.Scan(
			&orderInfo.OrderId,
			&orderInfo.Openid,
			&orderInfo.SchoolId,
			&orderInfo.Name,
			&orderInfo.Phone,
			&orderInfo.ParcelSize,
			&orderInfo.ParcelInfo,
			&orderInfo.SendAddr,
			&orderInfo.ReceiveAddr,
			&orderInfo.OrderTs,
			&orderInfo.ArrivalTs,
			&orderInfo.DispatchTs,
			&orderInfo.SendTs,
			&orderInfo.FinishTs,
			&orderInfo.Status,
			&orderInfo.Fee)
		orderList = append(orderList, &orderInfo)
	}
	return
}

func GetBoardOrderList(uin int64, schoolId int, orderType int) (orderList []*OrderInfo, err error) {
	//TODO 校验墙主

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	var sql string

	//墙主获取所有分配的订单
	if orderType == 5 {
		sql = fmt.Sprintf(`select orderId,openId,schoolId,name,phone,parcelSize,parcelInfo,sendAddr,receiveAddr,
				orderTs,arrivalTs,dispatchTs,sendTs,finishTs,status,fee,senderUid from express_orderInfo 
				where schoolId = %d and status in(2,3,4)`, schoolId)
	} else {
		//墙主不需要 已取消的订单  status = 1 取消
		sql = fmt.Sprintf(`select orderId,openId,schoolId,name,phone,parcelSize,parcelInfo,sendAddr,receiveAddr,
				orderTs,arrivalTs,dispatchTs,sendTs,finishTs,status,fee,senderUid from express_orderInfo 
				where schoolId = %d and status =%d`, schoolId, orderType)
	}

	log.Debugf("orderSql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	orderList = make([]*OrderInfo, 0)

	for rows.Next() {
		var orderInfo OrderInfo
		var senderUid int64
		rows.Scan(
			&orderInfo.OrderId,
			&orderInfo.Openid,
			&orderInfo.SchoolId,
			&orderInfo.Name,
			&orderInfo.Phone,
			&orderInfo.ParcelSize,
			&orderInfo.ParcelInfo,
			&orderInfo.SendAddr,
			&orderInfo.ReceiveAddr,
			&orderInfo.OrderTs,
			&orderInfo.ArrivalTs,
			&orderInfo.DispatchTs,
			&orderInfo.SendTs,
			&orderInfo.FinishTs,
			&orderInfo.Status,
			&orderInfo.Fee,
			&senderUid,
		)

		if senderUid > 0 {
			ui, err2 := GetSenderInfo(senderUid, schoolId)
			if err2 != nil {
				log.Debugf("GetBoardOrderList GetUserProfileInfo err, err:%+v", err2)
			}
			orderInfo.SenderInfo = &ui
		}

		log.Debugf("orderInfo:%+v", orderInfo)
		orderList = append(orderList, &orderInfo)
	}
	return
}

func GetSendOrderList(uin int64, schoolId int, orderType int) (orderList []*OrderInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select orderId,openId,schoolId,name,phone,parcelSize,parcelInfo,sendAddr,receiveAddr,
orderTs,arrivalTs,dispatchTs,sendTs,finishTs,status,fee,senderUid from express_orderInfo where schoolId = %d 
and status = %d and senderUid = %d`, schoolId, orderType, uin)

	log.Debugf("orderSql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	orderList = make([]*OrderInfo, 0)
	for rows.Next() {
		var orderInfo OrderInfo
		var senderUid int64
		rows.Scan(
			&orderInfo.OrderId,
			&orderInfo.Openid,
			&orderInfo.SchoolId,
			&orderInfo.Name,
			&orderInfo.Phone,
			&orderInfo.ParcelSize,
			&orderInfo.ParcelInfo,
			&orderInfo.SendAddr,
			&orderInfo.ReceiveAddr,
			&orderInfo.OrderTs,
			&orderInfo.ArrivalTs,
			&orderInfo.DispatchTs,
			&orderInfo.SendTs,
			&orderInfo.FinishTs,
			&orderInfo.Status,
			&orderInfo.Fee,
			&senderUid,
		)

		log.Debugf("senderUid:%d", senderUid)

		if senderUid > 0 {
			ui, err2 := GetSenderInfo(senderUid, schoolId)
			if err2 != nil {
				log.Debugf("GetBoardOrderList GetUserProfileInfo err, err:%+v", err2)
			}
			orderInfo.SenderInfo = &ui
		}

		orderList = append(orderList, &orderInfo)
	}
	return
}

func GetMyOrderList(openid string, schoolId int) (orderList []*OrderInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select orderId,openId,schoolId,name,phone,parcelSize,parcelInfo,
sendAddr,receiveAddr,orderTs,arrivalTs,dispatchTs,sendTs,finishTs ,status,fee from express_orderInfo 
where openid = '%s' and schoolId = %d`, openid, schoolId)

	log.Debugf("orderSql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	orderList = make([]*OrderInfo, 0)

	for rows.Next() {
		var orderInfo OrderInfo
		rows.Scan(
			&orderInfo.OrderId,
			&orderInfo.Openid,
			&orderInfo.SchoolId,
			&orderInfo.Name,
			&orderInfo.Phone,
			&orderInfo.ParcelSize,
			&orderInfo.ParcelInfo,
			&orderInfo.SendAddr,
			&orderInfo.ReceiveAddr,
			&orderInfo.OrderTs,
			&orderInfo.ArrivalTs,
			&orderInfo.DispatchTs,
			&orderInfo.SendTs,
			&orderInfo.FinishTs,
			&orderInfo.Status,
			&orderInfo.Fee)
		orderList = append(orderList, &orderInfo)
	}
	return
}
