package st

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

//检查是否我的好友
func CheckIsMyFriends(uin int64, fUins []int64) (res map[int64]int, err error) {

	res = make(map[int64]int)
	if uin == 0 || len(fUins) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	strs := ""
	for _, uid := range fUins {
		strs += fmt.Sprintf("%d,", uid)
	}
	strs = strs[:len(strs)-1]

	sql := fmt.Sprintf(`select friendUin from friends where uin = %d and friendUin in (%s)`, uin, strs)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var fUin int64
		rows.Scan(&fUin)

		res[fUin] = 1
	}

	return
}

//检查是否我的好友
func CheckIsMyFriend(uin int64, fUin int64) (res int, err error) {

	res = 0
	if uin == 0 || fUin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select friendUin from friends where uin = %d and friendUin = %d`, uin, fUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var fUin int64
		rows.Scan(&fUin)

		res = 1
		break
	}

	return
}

//检查是否我邀请过加好友
func CheckIsMyInvite(uin int64, fUin int64) (hasInvited int, err error) {

	hasInvited = 0

	if uin == 0 || fUin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select toUin from addFriendMsg where fromUin = %d and toUin = %d`, uin, fUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var t int64
		rows.Scan(&t)

		hasInvited = 1
	}

	return
}

//检查是否我邀请过加好友
func CheckIsMyInvites(uin int64, fUins []int64) (res map[int64]int, err error) {

	res = make(map[int64]int)
	if uin == 0 || len(fUins) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	strs := ""
	for _, uid := range fUins {
		strs += fmt.Sprintf("%d,", uid)
	}
	strs = strs[:len(strs)-1]

	sql := fmt.Sprintf(`select toUin from addFriendMsg where fromUin = %d and toUin in (%s)`, uin, strs)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var fUin int64
		rows.Scan(&fUin)

		res[fUin] = 1
	}

	return
}

//已经邀请但未接受的
func GetMyInviteUins(uin int64) (inviteUins []int64, err error) {

	inviteUins = make([]int64, 0)
	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select toUin from addFriendMsg where fromUin = %d `, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var fUin int64
		rows.Scan(&fUin)

		inviteUins = append(inviteUins, fUin)
	}

	return
}

//检查是否我邀请过发送短信
func CheckIsMyInvitesBySms(uin int64, phones []string) (res map[string]int, err error) {

	res = make(map[string]int)
	if uin == 0 || len(phones) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	strs := ""
	for _, phone := range phones {
		strs += fmt.Sprintf("\"%s\",", phone)
	}
	strs = strs[:len(strs)-1]

	sql := fmt.Sprintf(`select phone from inviteFriendSms where uin = %d and phone in (%s)`, uin, strs)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var phone string
		rows.Scan(&phone)

		res[phone] = 1
	}

	return
}

//检查对方是否邀请过加我为好友
func CheckIsInviteMe(uin int64, fUins []int64) (res map[int64]int, err error) {

	res = make(map[int64]int)
	if uin == 0 || len(fUins) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	strs := ""
	for _, uid := range fUins {
		strs += fmt.Sprintf("%d,", uid)
	}
	strs = strs[:len(strs)-1]

	sql := fmt.Sprintf(`select fromUin from addFriendMsg where toUin = %d and fromUin in (%s)`, uin, strs)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var fUin int64
		rows.Scan(&fUin)

		res[fUin] = 1
	}

	return
}

//检查对方与我的关系 好友 非好友 我已经申请加对方 对方已经申请加我
func GetUinStatusWithMe(uin, fUin int64) (status int, err error) {

	if uin == 0 || fUin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	//初始化为不是好友关系状态
	status = constant.ENUM_SNS_STATUS_NOT_FRIEND

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//检查是否好友
	sql := fmt.Sprintf(`select friendUin from friends where uin = %d and friendUin = %d`, uin, fUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	find := false

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)

		find = true
	}

	//好友
	if find {
		status = constant.ENUM_SNS_STATUS_IS_FRIEND
		return
	}

	//检查是否发送邀请过对方
	sql = fmt.Sprintf(`select msgId from addFriendMsg where fromUin = %d and toUin = %d`, uin, fUin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)

		find = true
	}

	//已经邀请过对方
	if find {
		status = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND
		return
	}

	//检查对方是否发送邀请给我
	sql = fmt.Sprintf(`select msgId from addFriendMsg where fromUin = %d and toUin = %d`, fUin, uin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)

		find = true
	}

	//对方已经邀请过我
	if find {
		status = constant.ENUM_SNS_STATUS_FRIEND_HAS_INVAITE_ME
		return
	}

	return
}

//检查对方与我的关系 好友 非好友 我已经申请加对方 对方已经申请加我
//非好友的情况下如果对方申请加我了， 同时返回邀请加我的消息ID, 搜索时使用可以直接点击接受
func GetUinStatusWithMe2(uin, fUin int64) (status int, msgId int64, err error) {

	if uin == 0 || fUin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	//初始化为不是好友关系状态 并且互相没有邀请
	status = constant.ENUM_SNS_STATUS_NOT_FRIEND
	msgId = 0

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//检查是否好友
	sql := fmt.Sprintf(`select friendUin from friends where uin = %d and friendUin = %d`, uin, fUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	find := false

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)

		find = true
	}

	//好友
	if find {
		status = constant.ENUM_SNS_STATUS_IS_FRIEND
		return
	}

	//检查是否发送邀请过对方
	sql = fmt.Sprintf(`select msgId from addFriendMsg where fromUin = %d and toUin = %d`, uin, fUin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		rows.Scan(&msgId)

		find = true
	}

	//已经邀请过对方
	if find {
		status = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND
		return
	}

	//检查对方是否发送邀请给我
	sql = fmt.Sprintf(`select msgId from addFriendMsg where fromUin = %d and toUin = %d`, fUin, uin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		rows.Scan(&msgId)

		find = true
	}

	//对方已经邀请过我
	if find {
		status = constant.ENUM_SNS_STATUS_FRIEND_HAS_INVAITE_ME
		return
	}

	return
}

//批量查询对方与我的关系 好友 非好友 我已经申请加对方 对方已经申请加我
func GetUinsStatusWithMe(uin int64, fUins []int64) (res map[int64]int, err error) {

	res = make(map[int64]int)
	if uin == 0 || len(fUins) == 0 {
		return
	}

	nUins := fUins

	//检查是否好友
	r, err := CheckIsMyFriends(uin, nUins)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for uid, _ := range r {
		res[uid] = constant.ENUM_SNS_STATUS_IS_FRIEND
	}

	//全部是好友
	if len(res) == len(fUins) {
		return
	}

	///.................................
	////检查是否邀请过对方
	nUins = make([]int64, 0)
	for _, uid := range fUins {
		if _, ok := res[uid]; ok {
			continue
		}

		nUins = append(nUins, uid)
	}

	if len(nUins) == 0 {
		return
	}

	r, err = CheckIsMyInvites(uin, nUins)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for uid, _ := range r {

		if _, ok := res[uid]; ok {
			continue
		}

		res[uid] = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND
	}

	//全部检查完成
	if len(res) == len(fUins) {
		return
	}

	///.................................
	//是否对方邀请过我
	nUins = make([]int64, 0)
	for _, uid := range fUins {
		if _, ok := res[uid]; ok {
			continue
		}

		nUins = append(nUins, uid)
	}

	if len(nUins) == 0 {
		return
	}

	r, err = CheckIsInviteMe(uin, nUins)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for uid, _ := range r {

		if _, ok := res[uid]; ok {
			continue
		}

		res[uid] = constant.ENUM_SNS_STATUS_FRIEND_HAS_INVAITE_ME
	}

	//全部完成
	if len(res) == len(fUins) {
		return
	}

	for _, uid := range fUins {

		if _, ok := res[uid]; ok {
			continue
		}

		res[uid] = constant.ENUM_SNS_STATUS_NOT_FRIEND
	}

	return
}
