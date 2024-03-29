package wxpublic

import (
	"common/mydb"
	//"io/ioutil"
	"common/constant"
	"common/rest"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type GetRandomRedPacketReq struct {
	UserOpenId string `schema:"userOpenId"`
	Phone      string `schema:"phone"`
}

type GetRandomRedPacketRsp struct {
	Amount int `schema:"amount"`
}

var (
	redPacketMux sync.Mutex
)

func doGetRandomRedPacket(req *GetRandomRedPacketReq, r *http.Request) (rsp *GetRandomRedPacketRsp, err error) {

	log.Debugf("GetRandomPacketNumReq %+v", req)

	amount, err := GetRandomRedPacket(req.UserOpenId, req.Phone)

	if err != nil {
		log.Errorf("openId %s, GetRandomRedPacket error %s", req.UserOpenId, err.Error())
		return
	}

	rsp = &GetRandomRedPacketRsp{amount}

	log.Errorf("openId %s, GetRandomRedPacket succ %d", req.UserOpenId, amount)
	return
}

func GetRandomRedPacket(openId, phone string) (amount int, err error) {

	if len(phone) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "phone invalid, length 0!")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	//查询明确领取过红包的
	//status = 0 未分配的红包
	//status = 1 已经分配, 但可能给用户发送红包失败
	//status = 2 已经分配，并且给用户发送红包成功
	sql := fmt.Sprintf(`select idx from redPacket where phone = "%s" and status >= 1`, phone)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		find = true
		break
	}

	if find {
		amount = 0
		//err = rest.NewAPIError(constant.E_USER_ALREADY_GET_RED_PACKET, openId+" user already get redpacket")
		//log.Errorf(err.Error())
		return
	}

	//下面的流程必须是在一个锁区间范围内执行，否则可能出现冲入的问题
	redPacketMux.Lock()
	defer redPacketMux.Unlock()

	sql = fmt.Sprintf(`select idx, amount from redPacket where status = 0 order by idx limit 1`)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	index := 0
	amount = 0
	for rows.Next() {
		rows.Scan(&index, &amount)
	}

	log.Errorf("user %s, phone %s, GetRandomRedPacket query from db, index %d, amount %d", openId, phone, index, amount)

	//没有红包了
	if amount == 0 {
		//err = rest.NewAPIError(constant.E_DB_QUERY, "db query index amount ret error")
		//log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()

	sql = fmt.Sprintf(`update redPacket set userOpenId = "%s", phone = "%s", status = 1, ts = %d where idx = %d`, openId, phone, ts, index)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("user %s, phone %s, GetRandomRedPacket succ ret, index %d, amount %d", openId, phone, index, amount)

	return
}

func HasGetRandomRedPacket(phone string) (hasGet int, err error) {

	hasGet = 1

	defer func() {
		log.Errorf("user phone %s, hasGetRandomPacket %d", phone, hasGet)
	}()

	if len(phone) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "phone invalid, length 0!")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	//查询明确领取过红包的
	//status = 0 未分配的红包
	//status = 1 已经分配, 但可能给用户发送红包失败
	//status = 2 已经分配，并且给用户发送红包成功
	sql := fmt.Sprintf(`select idx from redPacket where phone = "%s"`, phone)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		find = true
		break
	}

	if find {
		hasGet = 1
	} else {
		hasGet = 0
	}

	return
}

func UpdateRedPacketReceiveRecord(openId, phoneNum string) (err error) {

	log.Debugf("start UpdateRedPacketReceiveRecord openId:%s, phone %s", openId, phoneNum)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`update redPacket set status = %d where userOpenId = "%s" and phone = "%s"`, 2, openId, phoneNum)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	log.Debugf("end UpdateRedPacketReceiveRecord openId:%s, phone %s", openId, phoneNum)
	return
}
