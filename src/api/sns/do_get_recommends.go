package sns

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"common/token"
	"fmt"
	"net/http"
	"strconv"
	"svr/st"
)

type GetRecommendsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Type  int    `schema:"type"` //通讯录好友/好友的好友/同校好友
	//SubType int    `schema:"subType"`    //学校里面的子类别

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetRecommendsRsp struct {
	Total     int              `json:"total"`
	Recommend []*RecommendInfo `json:"friends"`
}

type RecommendInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`

	Grade      int    `json:"grade"`
	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"schoolName"`
	Phone      string `json:"phone"`

	Status        int    `json:"status"`        //0非好友 1好友 2已经邀请对方 3对方已经邀请我
	RecommendType int    `json:"recommendType"` //同校 通讯录已注册但是非好友 通讯录未注册
	RecommendDesc string `json:"recommendDesc"` //推荐描述
}

func (this *RecommendInfo) String() string {

	return fmt.Sprintf(`RecommendInfo{Uin:%d, NickName:%s, HeadImgUrl:%s, Gender:%d, Grade:%d, SchoolId:%d, SchoolType:%d, SchoolName:%s, Phone:%s, Status:%d, RecommendType:%d RecommendDesc:%s}`,
		this.Uin, this.NickName, this.HeadImgUrl, this.Gender, this.Grade, this.SchoolId, this.SchoolType, this.SchoolName, this.Phone, this.Status, this.RecommendType, this.RecommendDesc)
}

func doGetRecommends(req *GetRecommendsReq, r *http.Request) (rsp *GetRecommendsRsp, err error) {

	log.Debugf("uin %d, GetRecommendsReq %+v", req.Uin, req)

	//uuid 唯一标识设备, 用于查询通讯录
	uuid, err := token.GetUuidFromTokenString(req.Token, req.Ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt "+err.Error())
		log.Errorf("uin %d, GetRecommendsRsp error, %s", req.Uin, err.Error())
		return
	}

	total, friends, err := GetRecommends(req.Uin, req.Type, uuid, req.PageNum, req.PageSize)
	if err != nil {
		log.Errorf("uin %d, GetRecommendsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetRecommendsRsp{total, friends}

	log.Debugf("uin %d, GetRecommendsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetRecommends(uin int64, typ int, uuid int64, pageNum, pageSize int) (total int, friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

	if typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL || typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GRADE || typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_BOY || typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GIRL || typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_DEPT {

		total, friends, err = GetRecommendsFromSameSchool(uin, typ, pageNum, pageSize)

	} else if typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_REGISTED {

		total, friends, err = GetRecommendsFromAddrBookHasRegisted(uin, uuid, pageNum, pageSize)

	} else if typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_NOT_REGISTED {

		total, friends, err = GetRecommendsFromAddrBookNotRegisted(uin, uuid, pageNum, pageSize)

	} else if typ == constant.ENUM_RECOMMEND_FRIEND_TYPE_2DEGREE_FRIEND {

		total, friends, err = GetRecommendsFrom2DegreeFriends(uin, pageNum, pageSize)
	}

	return
}

func GetRecommendsFromSameSchool(uin int64, subType int, pageNum, pageSize int) (total int, friends []*RecommendInfo, err error) {

	log.Debugf("start GetRecommendsFromSameSchool uin:%d", uin)

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		return
	}

	//没有设置学校信息
	if ui.SchoolId == 0 {
		return
	}

	//获取我的好友UIN
	friendUins, err := st.GetMyFriendUins(uin)
	if err != nil {
		return
	}

	inviteUins, err := st.GetMyInviteUins(uin)
	if err != nil {
		return
	}

	strs := ""
	//排除好友列表
	for _, uid := range friendUins {
		strs += fmt.Sprintf("%d,", uid)
	}

	//排除已经邀请过的
	for _, uid := range inviteUins {
		strs += fmt.Sprintf("%d,", uid)
	}

	//要从同校中排除自己
	strs += fmt.Sprintf("%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//同校已经注册的
	conditions := fmt.Sprintf(`schoolId = %d and uin not in (%s)`, ui.SchoolId, strs)

	if subType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GRADE {

		conditions += fmt.Sprintf(` and grade = %d `, ui.Grade)

	} else if subType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_BOY {

		conditions += fmt.Sprintf(` and gender = 1 `)

	} else if subType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GIRL {

		conditions += fmt.Sprintf(` and gender = 2 `)

	} else if subType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_DEPT {

		if ui.DeptId > 0 {

			conditions += fmt.Sprintf(` and deptId = %d `, ui.DeptId)
		} else {
			log.Debugf("deptId is zero")
			return
		}
	}

	//过滤掉昵称为空的
	conditions += fmt.Sprintf(` and length(nickName) > 0`)

	sql := fmt.Sprintf(`select count(uin) from profiles where %s`, conditions)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	//总数为0
	if total == 0 {
		return
	}

	sql = fmt.Sprintf(`select uin, phone, nickName, headImgUrl, gender, grade, schoolId, schoolType, schoolName, deptId, deptName from profiles where %s order by abs(grade - %d) limit %d, %d`, conditions, ui.Grade, s, e)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	recommUins := make([]int64, 0)
	recommUinsMap := make(map[int64]*RecommendInfo)
	recommUinsMaptmp := make(map[int64]int)

	for rows.Next() {

		var fi RecommendInfo
		var deptId int
		var deptName string

		rows.Scan(&fi.Uin, &fi.Phone, &fi.NickName, &fi.HeadImgUrl, &fi.Gender, &fi.Grade, &fi.SchoolId, &fi.SchoolType, &fi.SchoolName, &deptId, &deptName)

		if len(fi.HeadImgUrl) > 0 {
			fi.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", fi.HeadImgUrl)
		}

		//初始为非好友状态
		fi.Status = constant.ENUM_SNS_STATUS_NOT_FRIEND
		fi.RecommendType = subType
		fi.RecommendDesc = fmt.Sprintf("同校%s", st.GetGradeDescBySchool(fi.SchoolType, fi.Grade))

		if subType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL {
			recommUinsMap[fi.Uin] = &fi
			recommUinsMaptmp[fi.Uin] = deptId
			recommUins = append(recommUins, fi.Uin)
		} else {
			friends = append(friends, &fi)
		}
	}

	// 同校同学按:同校同学院同年级 >> 同校同学院其他年级 >> 同校其他学院同年级 >>  同校其他学院其他年级
	if subType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL {
		sameDeptSameGradeUins := make([]int64, 0)
		sameDeptOtherGradeUins := make([]int64, 0)
		otherDeptSameGradeUins := make([]int64, 0)
		otherDeptOtherGradeUins := make([]int64, 0)

		for _, uid := range recommUins {
			if recommUinsMaptmp[uid] == ui.DeptId && recommUinsMap[uid].Grade == ui.Grade {
				sameDeptSameGradeUins = append(sameDeptSameGradeUins, uid)
			} else if recommUinsMaptmp[uid] == ui.DeptId {
				sameDeptOtherGradeUins = append(sameDeptOtherGradeUins, uid)
			} else if recommUinsMaptmp[uid] != ui.DeptId && recommUinsMap[uid].Grade == ui.Grade {
				otherDeptSameGradeUins = append(otherDeptSameGradeUins, uid)
			} else {
				otherDeptOtherGradeUins = append(otherDeptOtherGradeUins, uid)
			}
		}

		allUids := make([]int64, 0)
		allUids = append(allUids, sameDeptSameGradeUins...)
		allUids = append(allUids, sameDeptOtherGradeUins...)
		allUids = append(allUids, otherDeptSameGradeUins...)
		allUids = append(allUids, otherDeptOtherGradeUins...)

		for _, uid := range allUids {
			if info, ok := recommUinsMap[uid]; ok {
				friends = append(friends, info)
			}
		}
	}

	//是否已经发送加好友请求
	/*
		sts, err := CheckIsMyInvites(uin, recommUins)
		if err != nil{
			log.Error(err.Error())
			return
		}

		for i, fi := range friends{
			if _, ok := sts[fi.Uin]; ok{
				fi.Status = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND
				friends[i] = fi
			}
		}
	*/

	return
}

func GetRecommendsFromAddrBookHasRegisted(uin int64, uuid int64, pageNum, pageSize int) (total int, friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	//获取我的好友UIN
	friendUins, err := st.GetMyFriendUins(uin)
	if err != nil {
		return
	}

	inviteUins, err := st.GetMyInviteUins(uin)
	if err != nil {
		return
	}

	strs := ""
	//排除我的好友
	for _, uid := range friendUins {
		strs += fmt.Sprintf("%d,", uid)
	}

	//排除已经邀请过的
	for _, uid := range inviteUins {
		strs += fmt.Sprintf("%d,", uid)
	}
	strs += fmt.Sprintf("%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//查询通讯录中的非应用内的好友
	sql := fmt.Sprintf(`select count(friendUin) from addrBook where uuid = %d and friendUin > 0 and friendUin not in (%s)`, uuid, strs)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	//总数为0
	if total == 0 {
		return
	}

	//查询通讯录中的非应用内的好友
	sql = fmt.Sprintf(`select friendUin, friendName from addrBook where uuid = %d and friendUin > 0 and friendUin not in (%s) limit %d, %d`, uuid, strs, s, e)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	recommUins := make([]int64, 0)
	names := make(map[int64]string)
	for rows.Next() {
		var uid int64
		var name string
		rows.Scan(&uid, &name)

		recommUins = append(recommUins, uid)

		names[uid] = name
	}

	//获取这批用户的资料
	res, err := st.BatchGetUserProfileInfo(recommUins)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, uid := range recommUins {

		if v, ok := res[uid]; ok {

			var fi RecommendInfo

			fi.Uin = v.Uin
			fi.NickName = v.NickName
			fi.HeadImgUrl = v.HeadImgUrl
			fi.Gender = v.Gender
			//fi.Age        = v.Age
			fi.Grade = v.Grade
			fi.SchoolId = v.SchoolId
			fi.SchoolName = v.SchoolName
			fi.SchoolType = v.SchoolType

			fi.Phone = v.Phone

			//初始状态为非好友
			fi.Status = constant.ENUM_SNS_STATUS_NOT_FRIEND
			fi.RecommendType = constant.ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_REGISTED
			fi.RecommendDesc = "通讯录好友"

			if len(fi.NickName) == 0 {
				if len(names[fi.Uin]) > 0 {
					fi.NickName = names[fi.Uin]
				}
			}

			friends = append(friends, &fi)
		}
	}

	//检查是否已经邀请过对方加好友
	/*
		sts, err := CheckIsMyInvites(uin, recommUins)
		if err != nil{
			log.Error(err.Error())
			return
		}

		for i, fi := range friends{
			if _, ok := sts[fi.Uin]; ok{
				fi.Status = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND

				friends[i] = fi
			}
		}
	*/

	return
}

func GetRecommendsFromAddrBookNotRegisted(uin int64, uuid int64, pageNum, pageSize int) (total int, friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

	//全部从第一页开始计算
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
		log.Error(err.Error())
		return
	}

	//查询通讯录中的非应用内的好友
	sql := fmt.Sprintf(`select count(friendPhone) from addrBook where uuid = %d and friendUin = 0`, uuid)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	//总数为0
	if total == 0 {
		return
	}

	//查询通讯录中的非应用内的好友
	sql = fmt.Sprintf(`select friendName, friendPhone from addrBook where uuid = %d and friendUin = 0 limit %d, %d`, uuid, s, e)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	phones := make([]string, 0)

	for rows.Next() {

		var fi RecommendInfo

		rows.Scan(&fi.NickName, &fi.Phone)

		if len(fi.Phone) == 0 || len(fi.NickName) == 0 {
			continue
		}

		fi.Status = constant.ENUM_SNS_STATUS_NOT_INVITE_BY_SMS
		fi.RecommendType = constant.ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_NOT_REGISTED

		//通讯录未注册的显示信息为手机号
		fi.RecommendDesc = fi.Phone

		friends = append(friends, &fi)
		phones = append(phones, fi.Phone)
	}

	//检查是否已经邀请过对方加好友
	/*
		sts, err := CheckIsMyInvitesBySms(uin, phones)
		if err != nil{
			log.Error(err.Error())
			return
		}

		for i, fi := range friends{

			if _, ok := sts[fi.Phone]; ok{
				fi.Status = constant.ENUM_SNS_STATUS_HAS_INVAITE_BY_SMS
				friends[i] = fi
			}
		}
	*/

	return
}

func GetRecommendsFrom2DegreeFriends(uin int64, pageNum, pageSize int) (total int, friends []*RecommendInfo, err error) {

	//log.Debugf("uin %d, GetRecommendsFrom2DegreeFriends pageNum %d, pageSize %d", uin, pageNum, pageSize)

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_2DEGREE_FRIENDS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	total, err = app.ZCard(keyStr)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageNum * pageSize

	if total == 0 || s >= total {
		return
	}

	inviteUins, err := st.GetMyInviteUins(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//去除我的好友
	friendUins, err := st.GetMyFriendUins(uin)
	for _, uid := range friendUins {
		inviteUins = append(inviteUins, uid)
	}

	//获取所有的有共同好友的
	valsStr, err := app.ZRevRangeWithScores(keyStr, 0, -1)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if len(valsStr) == 0 {
		return
	}

	if len(valsStr)%2 != 0 {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScores return values cnt not even(2X)")
		log.Error(err.Error())
		return
	}

	var friendUin int64
	var commonCnt int64

	scores := make(map[int64]int64)

	members := make([]int64, 0)

	for i, valStr := range valsStr {

		if i%2 == 0 {
			friendUin, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScores value not interge")
				log.Error(err.Error())
				return
			}

		} else {

			commonCnt, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScores value not interge")
				log.Error(err.Error())
				return
			}

			if friendUin > 0 && commonCnt > 0 {
				scores[friendUin] = commonCnt

				members = append(members, friendUin)
			}
		}
	}

	//去掉已经邀请过的
	nmembers := make([]int64, 0)
	for _, member := range members {

		find := false
		for _, inviteUin := range inviteUins {
			if inviteUin == member {
				find = true
				break
			}
		}

		if !find {
			nmembers = append(nmembers, member)
		}
	}

	//过滤掉已经邀请过的
	total = len(nmembers)

	//第N页已经是空白
	if s >= len(nmembers) {
		return
	}

	//结束位置超过限制
	if e > len(nmembers) {
		e = len(nmembers)
	}

	//计算分页
	members = nmembers[s:e]

	//log.Debugf("uin %d, GetRecommendsFrom2DegreeFriends s %d, e %d, members cnt %d", uin, s, e, len(members))

	res, err := st.BatchGetUserProfileInfo(members)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, uid := range members {

		if v, ok := res[uid]; ok {

			var fi RecommendInfo

			fi.Uin = v.Uin
			fi.NickName = v.NickName
			fi.HeadImgUrl = v.HeadImgUrl
			fi.Gender = v.Gender
			//fi.Age        = v.Age
			fi.Grade = v.Grade
			fi.SchoolId = v.SchoolId
			fi.SchoolName = v.SchoolName
			fi.SchoolType = v.SchoolType

			fi.Phone = v.Phone

			//初始状态为非好友
			fi.Status = constant.ENUM_SNS_STATUS_NOT_FRIEND
			fi.RecommendType = constant.ENUM_RECOMMEND_FRIEND_TYPE_2DEGREE_FRIEND

			fi.RecommendDesc = fmt.Sprintf("%d位共同好友", scores[fi.Uin])

			//过滤掉昵称为空的
			if len(fi.NickName) == 0 {
				continue
			}

			friends = append(friends, &fi)
		}
	}

	return
}
