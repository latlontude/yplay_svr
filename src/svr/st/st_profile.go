package st

import (
	"common/constant"
	"common/env"
	"common/mydb"
	"common/rest"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type UserProfileInfo struct {
	Uin        int64  `json:"uin"`
	UserName   string `json:"userName"`
	Phone      string `json:"phone"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	Age        int    `json:"age"`
	Grade      int    `json:"grade"`

	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"schoolName"`
	DeptId     int    `json:"deptId"`   //大学的学院信息
	DeptName   string `json:"deptName"` //大学的学院信息

	Country        string `json:"country"`
	Province       string `json:"province"`
	City           string `json:"city"`
	Ts             int    `json:"ts"`
	EnrollmentYear int    `json:"enrollmentYear"` //入学年份
	Hometown       string `json:"hometown"`       //家乡

	GemCnt    int `json:"gemCnt"`
	FriendCnt int `json:"friendCnt"`
	NewsCnt   int `json:"newsCnt"`

	Type int    `json:"type"` //是否是白名单用户  0:非白名单  1:白名单
	Ext  string `json:"ext"`
	Src  int    `json:"src"` //来源  默认:0:同校  1:同城 2:其他

	IsAngelDays int `json:"isAngelDays"` //成为天使的时间

}

type ExtInfo struct {
	SecoendHeadImgUrl string `json:"secondHeadImgUrl"`
}

type UserProfileInfo2 struct {
	Uin        int64  `json:"uin"`
	UserName   string `json:"userName"`
	Phone      string `json:"phone"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	Age        int    `json:"age"`
	Grade      int    `json:"grade"`

	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"schoolName"`
	DeptId     int    `json:"deptId"`   //大学的学院信息
	DeptName   string `json:"deptName"` //大学的学院信息

	Country        string `json:"country"`
	Province       string `json:"province"`
	City           string `json:"city"`
	Ts             int    `json:"ts"`
	EnrollmentYear int    `json:"enrollmentYear"` //入学年份
	Hometown       string `json:"hometown"`       //家乡

	GemCnt    int    `json:"gemCnt"`
	FriendCnt int    `json:"friendCnt"`
	NewsCnt   int    `json:"newsCnt"`
	Src       int    `json:"src"`  //来源  默认:0:同校  1:同城 2:其他
	Type      int    `json:"type"` //是否是白名单用户  0:非白名单  1:白名单
	Ext       string `json:"ext"`
}

func (this *UserProfileInfo) String() string {

	return fmt.Sprintf(`UserProfileInfo2{Uin:%d, UserName:%s, Phone:%s, NickName:%s, HeadImgUrl:%s, Gender:%d,
Age:%d, Grade:%d, SchoolId:%d, SchoolType:%d, SchoolName:%s, DeptId:%d, DeptName:%s, GemCnt:%d, FriendCnt:%d,
NewsCnt:%d,Src:%d,type:%d,ext:%s}`,
		this.Uin, this.UserName, this.Phone, this.NickName, this.HeadImgUrl, this.Gender,
		this.Age, this.Grade, this.SchoolId, this.SchoolType, this.SchoolName, this.DeptId,
		this.DeptName, this.GemCnt, this.FriendCnt, this.NewsCnt, this.Src, this.Type, this.Ext)
}

func (this *UserProfileInfo2) String() string {

	return fmt.Sprintf(`UserProfileInfo2{Uin:%d, UserName:%s, Phone:%s, NickName:%s, HeadImgUrl:%s, Gender:%d,
Age:%d, Grade:%d, SchoolId:%d, SchoolType:%d, SchoolName:%s, DeptId:%d, DeptName:%s, GemCnt:%d, FriendCnt:%d,
NewsCnt:%d,Src:%d,type:%d,ext:%s}`,
		this.Uin, this.UserName, this.Phone, this.NickName, this.HeadImgUrl, this.Gender,
		this.Age, this.Grade, this.SchoolId, this.SchoolType, this.SchoolName, this.DeptId,
		this.DeptName, this.GemCnt, this.FriendCnt, this.NewsCnt, this.Src, this.Type, this.Ext)
}

// func CopyProfileInfo2ProfileInfo2(info *UserProfileInfo, info2 *UserProfileInfo2) {
// 	if info == nil || info2 == nil {
// 		return
// 	}

// 	info2.Uin = info.Uin
// 	info2.UserName = info.UserName
// 	info2.Phone = info.Phone
// 	info2.NickName = info.NickName
// 	info2.HeadImgUrl = info.HeadImgUrl
// 	info2.Gender = info.Gender
// 	info2.Age = info.Age
// 	info2.Grade = info.Grade

// 	info2.SchoolId = info.SchoolId
// 	info2.SchoolType = info.SchoolType
// 	info2.SchoolName = info.SchoolName

// 	info2.Country = info.Country
// 	info2.Province = info.Province
// 	info2.City = info.City
// 	info2.Ts = info.Ts

// 	return
// }

func GetUserProfileInfo(uin int64) (info *UserProfileInfo, err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}
	sql := fmt.Sprintf(`select uin, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType,schoolName, 
		deptId, deptName, country, province, city ,enrollmentYear,hometown from profiles where uin = %d`, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	find := false

	info = &UserProfileInfo{}

	for rows.Next() {

		rows.Scan(
			&info.Uin,
			&info.UserName,
			&info.Phone,
			&info.NickName,
			&info.HeadImgUrl,
			&info.Gender,
			&info.Age,
			&info.Grade,
			&info.SchoolId,
			&info.SchoolType,
			&info.SchoolName,
			&info.DeptId,
			&info.DeptName,
			&info.Country,
			&info.Province,
			&info.City,
			&info.EnrollmentYear,
			&info.Hometown)

		if len(info.HeadImgUrl) > 0 {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", info.HeadImgUrl)
		} else {
			//info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/defaultHeader.png")
		}

		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_USER_NOT_EXIST, "user not exist")
		return
	}

	//判断白名单 写到Type里面s
	whiteList := strings.Split(env.Config.WhiteList.Phones, ",") //内部测试手机号
	isWhitePhone := false

	for _, value := range whiteList {
		if value == info.Phone {
			isWhitePhone = true
			break
		}
	}

	var extInfo ExtInfo
	if isWhitePhone {
		info.Type = 1
		headImg := getSecondHeadImgUrl(inst, info.Phone)
		if len(headImg) > 0 {
			extInfo.SecoendHeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", headImg)
		}

		data, err1 := json.Marshal(&extInfo)
		if err1 != nil {
			log.Errorf(err1.Error())
		}
		dataStr := string(data)
		info.Ext = dataStr
	}

	return
}

func getSecondHeadImgUrl(inst *sql.DB, phone string) (secoendHeadImgUrl string) {
	sql := fmt.Sprintf(`select headImgUrl from secondHeadImgUrl where phone = %s`, phone)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&secoendHeadImgUrl)
	}
	secoendHeadImgUrl = strings.Trim(secoendHeadImgUrl, " \t\r\n")
	return
}

func BatchGetUserProfileInfo(uins []int64) (res map[int64]*UserProfileInfo, err error) {

	res = make(map[int64]*UserProfileInfo)

	if len(uins) == 0 {
		return
	}

	str := ``
	for i, uin := range uins {

		if i != len(uins)-1 {
			str += fmt.Sprintf(`%d,`, uin)
		} else {
			str += fmt.Sprintf(`%d`, uin)
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select uin, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType, schoolName, deptId, deptName, country, province, city ,enrollmentYear,hometown from profiles where uin in (%s)`, str)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info UserProfileInfo

		rows.Scan(
			&info.Uin,
			&info.UserName,
			&info.Phone,
			&info.NickName,
			&info.HeadImgUrl,
			&info.Gender,
			&info.Age,
			&info.Grade,
			&info.SchoolId,
			&info.SchoolType,
			&info.SchoolName,
			&info.DeptId,
			&info.DeptName,
			&info.Country,
			&info.Province,
			&info.City,
			&info.EnrollmentYear,
			&info.Hometown)

		if len(info.HeadImgUrl) > 0 {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", info.HeadImgUrl)
		} else {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/defaultHeader.png")
		}

		res[info.Uin] = &info
	}

	return
}

func GetUserProfileInfo2(uin int64) (info *UserProfileInfo2, err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}
	sql := fmt.Sprintf(`select uin, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType,schoolName, deptId, deptName, country, province, city ,enrollmentYear,hometown from profiles where uin = %d`, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	info = &UserProfileInfo2{}

	find := false

	for rows.Next() {

		rows.Scan(
			&info.Uin,
			&info.UserName,
			&info.Phone,
			&info.NickName,
			&info.HeadImgUrl,
			&info.Gender,
			&info.Age,
			&info.Grade,
			&info.SchoolId,
			&info.SchoolType,
			&info.SchoolName,
			&info.DeptId,
			&info.DeptName,
			&info.Country,
			&info.Province,
			&info.City,
			&info.EnrollmentYear,
			&info.Hometown)

		if len(info.HeadImgUrl) > 0 {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", info.HeadImgUrl)
		} else {
			//info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/defaultHeader.png")
		}

		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_USER_NOT_EXIST, "user not exist")
		return
	}

	sql = fmt.Sprintf(`select uin, statField, statValue from userStat where uin = %d`, uin)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uid int64
		var field int
		var value int

		rows.Scan(&uid, &field, &value)

		if field == constant.ENUM_USER_STAT_GEM_CNT {
			info.GemCnt = value
		}

		if field == constant.ENUM_USER_STAT_FRIEND_CNT {
			info.FriendCnt = value
		}
	}

	//判断白名单 写到Type里面s
	whiteList := strings.Split(env.Config.WhiteList.Phones, ",") //内部测试手机号

	isWhitePhone := false

	for _, value := range whiteList {
		if value == info.Phone {
			isWhitePhone = true
			break
		}
	}

	var extInfo ExtInfo
	if isWhitePhone {
		info.Type = 1
		headImg := getSecondHeadImgUrl(inst, info.Phone)
		if len(headImg) > 0 {
			extInfo.SecoendHeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", headImg)
			log.Debugf("phone:%s  ,  url:%s", info.Phone, extInfo.SecoendHeadImgUrl)
		}

		data, err1 := json.Marshal(&extInfo)
		if err1 != nil {
			log.Errorf(err1.Error())
		}
		dataStr := string(data)
		info.Ext = dataStr
	}

	isBoardOwner, isSender, err := CheckExpressUser(uin, info.SchoolId)
	log.Debugf("checkUser:uin:%d,isBoardOwner:%t,isSender:%t", uin, isBoardOwner, isSender)
	if isBoardOwner == true || isSender == true {
		info.Type = 2
	}

	return
}

func CheckExpressUser(uin int64, schoolId int) (isBoardAdmin bool, isSender bool, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	{
		//判断是不是墙主
		sql := fmt.Sprintf(`select ownerUid from v2boards where schoolId = %d`, schoolId)
		rows, err1 := inst.Query(sql)
		defer rows.Close()
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
			log.Error(err)
			return
		}
		isBoardAdmin = false
		for rows.Next() {
			var uid int64
			rows.Scan(&uid)
			if uid == uin {
				isBoardAdmin = true
			}
		}
	}

	{
		//判断是不是跑腿者
		sql := fmt.Sprintf(`select uid from express_senderList where schoolId = %d`, schoolId)
		rows, err1 := inst.Query(sql)
		defer rows.Close()
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
			log.Error(err)
			return
		}
		isSender = false
		for rows.Next() {
			var uid int64
			rows.Scan(&uid)
			if uid == uin {
				isSender = true
			}
		}
	}

	return
}

func BatchGetUserProfileInfo2(uins []int64) (res map[int64]*UserProfileInfo2, err error) {

	res = make(map[int64]*UserProfileInfo2)

	if len(uins) == 0 {
		return
	}

	str := ``
	for i, uin := range uins {

		if i != len(uins)-1 {
			str += fmt.Sprintf(`%d,`, uin)
		} else {
			str += fmt.Sprintf(`%d`, uin)
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select uin, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType, schoolName, deptId, deptName, country, province, city from profiles where uin in (%s)`, str)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info UserProfileInfo2

		rows.Scan(
			&info.Uin,
			&info.UserName,
			&info.Phone,
			&info.NickName,
			&info.HeadImgUrl,
			&info.Gender,
			&info.Age,
			&info.Grade,
			&info.SchoolId,
			&info.SchoolType,
			&info.SchoolName,
			&info.DeptId,
			&info.DeptName,
			&info.Country,
			&info.Province,
			&info.City)

		if len(info.HeadImgUrl) > 0 {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", info.HeadImgUrl)
		} else {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/defaultHeader.png")
		}

		res[info.Uin] = &info
	}

	sql = fmt.Sprintf(`select uin, statField, statValue from userStat where uin in (%s)`, str)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uid int64
		var field int
		var value int

		rows.Scan(&uid, &field, &value)

		if field == constant.ENUM_USER_STAT_GEM_CNT {

			if _, ok := res[uid]; ok {
				res[uid].GemCnt = value
			}
		}

		if field == constant.ENUM_USER_STAT_FRIEND_CNT {

			if _, ok := res[uid]; ok {
				res[uid].GemCnt = value
			}
		}
	}

	return
}
