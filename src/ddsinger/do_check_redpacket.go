package ddsinger

import (
	"common/mydb"
	//"io/ioutil"
	"common/constant"
	"common/rest"
	"fmt"
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

	sql := fmt.Sprintf(`select id from ddSingerRedPacket where phone = "%s"`, phone)

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

func RecordSendRedPacket(openId string, phone string) (cnt int, err error) {

	log.Debugf("start RecordSendRedPacket openId:%s, phone:%s", openId, phone)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select uin from profiles where phone = "%s"`, phone)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	var uid int64
	for rows.Next() {
		rows.Scan(&uid)
	}

	if uid == 0 {
		log.Debugf(" not exist user which phone = %s ", phone)
		return
	}

	sql = fmt.Sprintf(`select count(id) as cnt from ddcallForSingers where singerId = 2 and uin = %d and type = 8 `, uid)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&cnt)
	}

	if cnt == 0 {
		log.Debugf(" call cnt is zero uin :%d", uid)
		return
	}
	stmt, err := inst.Prepare(`insert into ddSingerRedPacket values(?, ?, ?, ?, ?,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()
	status := 0

	_, err = stmt.Exec(0, openId, phone, cnt, status, ts)
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

	sql := fmt.Sprintf(`update ddSingerRedPacket set status = %d where openId = "%s" and phone = "%s"`, 1, openId, phoneNum)

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
