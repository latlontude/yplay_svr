package chengyuan

import (
	"common/mydb"
	//"io/ioutil"
	"common/constant"
	"common/rest"
	"fmt"
	"math/rand"
	"time"
)

func HasGetRedPacket(phone string) (hasGet int, err error) {

	hasGet = 1

	defer func() {
		log.Errorf("user phone %s, hasGetPacket %d", phone, hasGet)
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

	//查询是否领取过红包的
	//status = 0 已经发送
	//status = 1 已经成功

	sql := fmt.Sprintf(`select id from cyRedPacket where phone = "%s"`, phone)

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

func RecordSendRedPacket(openId string, phone string) (money int, err error) {

	log.Debugf("start RecordSendRedPacket openId:%s, phone:%s", openId, phone)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into cyRedPacket values(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	rand.Seed(time.Now().Unix())
	money = rand.Intn(10) + 1
	ts := time.Now().Unix()
	status := 0

	_, err = stmt.Exec(0, openId, phone, money, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end RecordSendRedPacket openId:%s, phone:%s", openId, phone)
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

	sql := fmt.Sprintf(`update cyRedPacket set status = %d where openId = "%s" and phone = "%s"`, 1, openId, phoneNum)

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
