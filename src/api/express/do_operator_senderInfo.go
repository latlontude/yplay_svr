//墙主查看 该学校的跑腿者

package express

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
	"time"
)

type SenderInfo struct {
	SchoolId     int    `json:"schoolId"`
	Uin          int64  `json:"uin"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	HeadImageUrl string `json:"headImageUrl"`
}

//墙主获取跑腿者列表
func GetSenderList(uin int64, schoolId int) (senderList []*SenderInfo, err error) {
	log.Debugf("uin:%d", uin)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//墙主不需要 已取消的订单  status = 1 取消
	sql := fmt.Sprintf(`select uid,name,phone ,schoolId from express_senderList where schoolId = %d and status = 0`, schoolId)

	log.Debugf("express_senderList sql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	senderList = make([]*SenderInfo, 0)

	for rows.Next() {
		var senderInfo SenderInfo
		rows.Scan(&senderInfo.Uin, &senderInfo.Name, &senderInfo.Phone, &senderInfo.SchoolId)

		if senderInfo.Uin > 0 {
			ui, err2 := st.GetUserProfileInfo(senderInfo.Uin)
			if err2 != nil {
				log.Debugf("GetBoardOrderList GetUserProfileInfo err, err:%+v", err2)
			}
			senderInfo.HeadImageUrl = ui.HeadImgUrl
		}
		senderList = append(senderList, &senderInfo)
	}
	return
}

func AddSenderInfo(uin int64, schoolId int, userName string, name string, phone string) (senderInfo SenderInfo, err error) {
	log.Debugf("uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//墙主不需要 已取消的订单  status = 1 取消
	sql := fmt.Sprintf(`select uin,headImgUrl from profiles where schoolId = %d and userName = '%s'`, schoolId, userName)

	log.Debugf("sendInfo:%s", sql)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()
	var uid int64
	var headImageUrl string
	for rows.Next() {
		rows.Scan(&uid, &headImageUrl)
	}

	if uid == 0 {
		log.Debugf("query uid error :%s", sql)
		err = rest.NewAPIError(constant.E_DB_QUERY, "user not exist")
		return
	}

	stmt, err := inst.Prepare(`insert into express_senderList(uid,name,phone,schoolId,ts,status) values(?,?,?,?,?,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	now := time.Now().Unix()
	res, err := stmt.Exec(uid, name, phone, schoolId, now, 0)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PRIMARY_KEY, err.Error())
		log.Error(err.Error())
		return
	}
	_, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	if uid == 0 {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	senderInfo.Uin = uid
	senderInfo.HeadImageUrl = headImageUrl
	senderInfo.Name = name
	senderInfo.Phone = phone
	senderInfo.SchoolId = schoolId

	return
}

func GetSenderInfo(uin int64, schoolId int) (senderInfo SenderInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select name,phone from express_senderList where schoolId = %d and uid = %d`, schoolId, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()
	var name, phone string
	for rows.Next() {
		rows.Scan(&name, &phone)
	}

	senderInfo.Uin = uin
	senderInfo.Name = name
	senderInfo.Phone = phone
	senderInfo.SchoolId = schoolId

	return
}
