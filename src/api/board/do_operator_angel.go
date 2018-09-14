package board

import (
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
	"time"
)

func AddAngelInAdmin(uin int64, boardId int, labelId int, AngelUin int64, adminType int ) (err error) {

	if uin < 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	//先判断是不是已经加入了admin表
	sql := fmt.Sprintf(`select count(uin) from experience_admin where boardId = %d and uin = %d`, boardId, AngelUin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		rows.Scan(&count)
	}

	if count == 1 {
		log.Debugf("have joined admin :uin:%d,angelUin:%d", uin, AngelUin)
		return
	}
	stmt, err := inst.Prepare(`insert into experience_admin(id, boardId, labelId,uin,ts ,type) values(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	_, err = stmt.Exec(0, boardId, labelId, AngelUin, ts ,adminType)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

//卸任天使
func DelAngelFromAdmin(uin int64, boardId int, AngelUin int64) (err error) {

	permissionList, err := CheckOperatorAngelPermission(uin, boardId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	log.Debugf("permissionList:%v", permissionList)

	//无权限
	if !permissionList["board"] && !permissionList["self"] && !permissionList["superAdmin"] {
		err = rest.NewAPIError(constant.E_PERMISSION, "")
		log.Error(err)
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`delete from experience_admin where boardId = %d and uin = %d`, boardId, AngelUin)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	sql = fmt.Sprintf(`update invite_angel set status = 1 where toUin = %d and boardId = %d`, AngelUin, boardId)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//卸任天使 发送push
	if permissionList["board"] {
		go v2push.SendDeleteAngelByBigAngelPush(uin, AngelUin)
	} else if permissionList["self"] {
		boardInfo, err2 := GetBoardInfoByBoardId(uin, boardId)
		if err2 != nil {
			log.Debugf("get boardInfo error uin:%d,angelUin:%d", uin, AngelUin)
			return
		}
		go v2push.SendDeleteAngelBySelfPush(uin, boardInfo.OwnerInfo.Uin)
	}

	return
}

//禅让墙主
func DemiseBigAngel(uin int64, toUin int64, boardId int) (err error) {

	log.Debugf("uin:%d,toUin:%d,boardId:%d", uin, toUin, boardId)

	if toUin == 0 {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "toUin is zero")
		log.Error(err)
		return
	}

	permissionList, err := CheckOperatorAngelPermission(uin, boardId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//无权限
	if !permissionList["board"] {
		err = rest.NewAPIError(constant.E_PERMISSION, "you don't have permission")
		log.Error(err)
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//查找所有小天使
	sql := fmt.Sprintf(`select uin from experience_admin where boardId = %d and type = 0`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	smallAngelList := make([]int64, 0)
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		smallAngelList = append(smallAngelList, uid)
	}

	//更新admin表状态
	sqlArr := make([]string, 0)
	sqlArr = append(sqlArr, fmt.Sprintf(`update experience_admin set type = 0 where boardId = %d and uin = %d`, boardId, uin))
	sqlArr = append(sqlArr, fmt.Sprintf(`update experience_admin set type = 1 where boardId = %d and uin = %d`, boardId, toUin))

	//还需要更新 v2boards表里 ownerUid(墙主)
	sqlArr = append(sqlArr, fmt.Sprintf(`update v2boards set ownerUid = %d where boardId = %d`, toUin, boardId))

	//删除内存缓存
	delete(boardMap, boardId)

	err = mydb.Exec(inst, sqlArr)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	for _, smallAngelUin := range smallAngelList {
		go v2push.SendDemiseAngelPush(uin, toUin, smallAngelUin)
	}
	return
}

//校验是否有操作天使的权限

func CheckOperatorAngelPermission(uin int64, boardId int) (permissionList map[string]bool, err error) {

	permissionList = make(map[string]bool, 0)

	//1.超级管理员  100001
	if uin == 100001 {
		permissionList["superAdmin"] = true
		return
	}

	//2.墙主
	boardInfo, ok := boardMap[boardId]
	if !ok {
		boardInfoTmp, err1 := GetBoardInfoByBoardId(uin, boardId)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
			log.Errorf(err.Error())
			return
		}
		boardInfo = &boardInfoTmp
	}

	if uin == boardInfo.OwnerInfo.Uin {
		permissionList["board"] = true
		return
	}

	//3.自己是管理员
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select uin from experience_admin where boardId = %d`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		if uid == uin {
			permissionList["self"] = true
		}
	}
	return
}

func InviteAngel(uin int64, AngelUin int64, boardId int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//1.判断是不是墙主
	permissionList, err := CheckOperatorAngelPermission(uin, boardId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	if !permissionList["board"] {
		err = rest.NewAPIError(constant.E_PERMISSION, "you don't have permission")
		log.Error(err)
		return
	}

	//2.先判断是不是天使团成员 不在列表才可以邀请
	sql := fmt.Sprintf(`select uin from experience_admin where boardId = %d`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	isInAngelList := false
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		if uid == AngelUin {
			isInAngelList = true
		}
	}

	if isInAngelList {
		err = rest.NewAPIError(constant.E_DB_QUERY, "is angel now ! , cann't repeat invite!")
		log.Errorf(err.Error())
		return
	}

	//3.判断是不是同一个学校的?

	myInfo, err := st.GetUserProfileInfo(uin)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, "GetUserInfo error")
		log.Errorf(err.Error())
		return
	}
	angelInfo, err := st.GetUserProfileInfo(AngelUin)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, "GetUserInfo error")
		log.Errorf(err.Error())
		return
	}

	//不是同一个学校的
	if myInfo.SchoolId != angelInfo.SchoolId && uin != 100001 {
		log.Errorf("different school ,uin:%d,angelUin:%d", uin, AngelUin)
		return
	}

	//3.插入表
	stmt, err := inst.Prepare(`insert into invite_angel values(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()
	sendTs := time.Now().Unix()
	res, err := stmt.Exec(0, boardId, uin, AngelUin, 0, sendTs, 0)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//获取新增数据id
	msgId, err := res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//4.发送通知
	go v2push.SendInviteAngelPush(uin, AngelUin, boardId, msgId)
	return
}

func AcceptAngel(uin int64, boardId int, msgId int) (err error) {

	//1.查询邀请记录
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	//2.先判断是不是天使团成员 不在列表才可以邀请
	sql := fmt.Sprintf(`select fromUin,toUin,status from invite_angel where msgId = %d and boardId = %d`, msgId, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	flag := false
	status := 0
	var fromUin, toUin int64
	for rows.Next() {
		rows.Scan(&fromUin, &toUin, &status)
		//有此记录
		if uin == toUin {
			flag = true
		}
	}

	//有此记录
	if flag == true {
		if status == 0 {
			//更新invite表状态
			sql = fmt.Sprintf(`update invite_angel set status = 1 where msgId = %d and boardId = %d`, msgId, boardId)
			_, err = inst.Exec(sql)
			if err != nil {
				err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
				log.Error(err.Error())
				return
			}
			//加入admin表
			err = AddAngelInAdmin(uin, boardId, 0, uin , 0)
			if err != nil {
				log.Errorf("add angel err , uin:%d err:%+v", uin, err.Error())
				return
			}
		} else {
			err = rest.NewAPIError(constant.E_ACCEPT_ANGEL_EXPIRE, "邀请已过期")
			log.Errorf(err.Error())
			return
		}
	} else {
		err = rest.NewAPIError(constant.E_ACCEPT_ANGEL_NO_RECORD, "无邀请记录")
		log.Errorf(err.Error())
		return
	}

	//接受邀请发送 push  toUin 小天使 fromUin 主天使
	go v2push.SendAcceptAngelPush(toUin, fromUin)

	return
}
