package st

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

type FriendInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	//	Age        int    `json:"age"`
	Grade      int    `json:"grade"`
	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"schoolName"`
	DeptId     int    `json:"deptId"`
	DeptName   string `json:"deptName"`

	Ts int `json:"ts"` //成为好友的时间
}

func (this *FriendInfo) String() string {

	return fmt.Sprintf(`FriendInfo{Uin:%d, NickName:%s, HeadImgUrl:%s, Gender:%d, Grade:%d, SchoolId:%d, SchoolType:%d, SchoolName:%s, Ts:%d}`,
		this.Uin, this.NickName, this.HeadImgUrl, this.Gender, this.Grade, this.SchoolId, this.SchoolType, this.SchoolName, this.Ts)
}

func GetMyFriends(uin int64, pageNum, pageSize int) (total int, friends []*FriendInfo, err error) {

	friends = make([]*FriendInfo, 0)

	if uin == 0 {
		return
	}

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(friendUin) from friends where uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	sql = fmt.Sprintf(`select friendUin, ts from friends where uin = %d order by ts desc limit %d, %d`, uin, s, e)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	uins := make([]int64, 0)
	uinsTs := make(map[int64]int)

	for rows.Next() {

		var uid int64
		var ts int
		rows.Scan(&uid, &ts)

		if uid == 0 || uid == uin {
			continue
		}

		uins = append(uins, uid)
		uinsTs[uid] = ts
	}

	res, err := BatchGetUserProfileInfo(uins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uid := range uins {

		if v, ok := res[uid]; ok {

			var friend FriendInfo

			friend.Uin = v.Uin
			friend.NickName = v.NickName
			friend.HeadImgUrl = v.HeadImgUrl
			friend.Gender = v.Gender
			friend.Grade = v.Grade
			friend.SchoolId = v.SchoolId
			friend.SchoolType = v.SchoolType
			friend.SchoolName = v.SchoolName
			friend.DeptId = v.DeptId
			friend.DeptName = v.DeptName
			friend.Ts = uinsTs[v.Uin]

			friends = append(friends, &friend)
		}
	}

	return
}

func GetMyFriendsCnt(uin int64) (total int, err error) {
	log.Debugf("start GetMyFriendsCnt uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(friendUin) from friends where uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	log.Debugf("end GetMyFriendsCnt total:%d", total)
	return
}

func GetAllMyFriends(uin int64) (friends map[int64]*FriendInfo, err error) {

	friends = make(map[int64]*FriendInfo)

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select friendUin, ts from friends where uin = %d order by ts desc`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	uins := make([]int64, 0)
	uinsTs := make(map[int64]int)
	for rows.Next() {

		var uid int64
		var ts int
		rows.Scan(&uid, &ts)

		if uid == 0 || uid == uin {
			continue
		}

		uins = append(uins, uid)
		uinsTs[uid] = ts
	}

	res, err := BatchGetUserProfileInfo(uins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uid := range uins {

		if v, ok := res[uid]; ok {

			var friend FriendInfo

			friend.Uin = v.Uin
			friend.NickName = v.NickName
			friend.HeadImgUrl = v.HeadImgUrl
			friend.Gender = v.Gender
			friend.Grade = v.Grade
			friend.SchoolId = v.SchoolId
			friend.SchoolType = v.SchoolType
			friend.SchoolName = v.SchoolName
			friend.Ts = uinsTs[v.Uin]

			friends[v.Uin] = &friend
		}
	}

	return
}

func GetMyFriendUins(uin int64) (uins []int64, err error) {

	uins = make([]int64, 0)

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	sql := fmt.Sprintf(`select friendUin from friends where uin = %d`, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uid int64
		rows.Scan(&uid)
		if uid == 0 || uid == uin {
			continue
		}

		uins = append(uins, uid)
	}

	return
}

func IsFriend(uin int64, friendUin int64) (isFriend int, err error) {

	if uin == 0 || friendUin == 0 {
		isFriend = 0
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	isFriend = 0

	sql := fmt.Sprintf(`select status from friends where uin = %d and friendUin = %d`, uin, friendUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)
		isFriend = 1
		break
	}

	return
}

func GetFriendListVer(uin int64) (ver int64, err error) {

	//默认版本号1，不要从0开始
	ver = 1

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select ver from friendListVer where uin = %d `, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&ver)
	}

	return
}
