package express

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	MCH_ID  = "1498480632"
	WXAPPID = "wxc6a993bffd64bb9a"
	KEY     = "4sbBcRjiGtWMSsOt56JYR1knwkTWe5TU" //key为商户平台设置的密钥key

	APIMap = httputil.APIMap{
		"/login":            auth.Apify(doGetCode),          //登陆
		"/order":            auth.Apify(doOrder),            //下单
		"/orderCallBack":    auth.Apify(doOrderCallBack),    //支付回调
		"/getFee":           auth.Apify(doGetFee),           //获取跑单费用
		"/getSchoolList":    auth.Apify(doGetSchoolList),    //获取学校列表
		"/getOrderId":       auth.Apify(doGetOrderId),       //获取学校列表
		"/getMyOrderList":   auth.Apify(doGetMyOrderList),   //我的订单列表
		"/getSendOrderList": auth.Apify(doGetSendOrderList), //跑单列表
		"/updateAnOrder":    auth.Apify(doUpdateAnOrder),    //跑腿者更改订单状态

		"/dispatchOrder":     auth.Apify(doDispatchOrder),     //派单
		"/getBoardOrderList": auth.Apify(doGetBoardOrderList), //墙主看到的订单列表

		"/getSenderList": auth.Apify(doGetSenderList), //墙主查看跑腿者列表
		"/addSenderInfo": auth.Apify(doAddSenderInfo), //添加跑腿者

		"/getAddrList": auth.Apify(doGetAddrList), //学校快递点

		"/checkUser": auth.Apify(doCheckUser), //学校快递点
	}
	log = env.NewLogger("express")

	EXPRESS_DISPATCH = "express_dispatch"
)
