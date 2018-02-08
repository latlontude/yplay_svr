package sns

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"common/token"
	"fmt"
	"math/rand"
	"net/http"
	"svr/st"
)

type GetRandomRecommendsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Type  int    `schema:"type"` //通讯录好友/好友的好友/同校好友
}

type GetRandomRecommendsRsp struct {
	Recommend []*RecommendInfo `json:"friends"`
}

func doGetRandomRecommends(req *GetRandomRecommendsReq, r *http.Request) (rsp *GetRandomRecommendsRsp, err error) {

	log.Debugf("uin %d, GetRandomRecommendsReq %+v", req.Uin, req)

	//uuid 唯一标识设备, 用于查询通讯录
	uuid, err := token.GetUuidFromTokenString(req.Token, req.Ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt "+err.Error())
		log.Errorf("uin %d, GetRandomRecommendsRsp error, %s", req.Uin, err.Error())
		return
	}

	//随机从通讯录和同校中拉取注册非好友
	friends, err := GetRandomRecommends(req.Uin, uuid, 2)
	if err != nil {
		log.Errorf("uin %d, GetRandomRecommendsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetRandomRecommendsRsp{friends}

	log.Debugf("uin %d, GetRandomRecommendsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetRandomRecommends(uin int64, uuid int64, cnt int) (friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 || uuid == 0 {
		return
	}

	inviteUins, err := st.GetMyInviteUins(uin)
	if err != nil {
		return
	}

	_, f1, err := GetRandomRecommendsFromSameSchoolExceptInvited(uin, inviteUins, cnt)
	if err != nil {
		log.Error(err.Error())
		return
	}

	_, f2, err := GetRandomRecommendsFromAddrBookHasRegistedExceptInvited(uin, uuid, inviteUins, cnt)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//去重
	for _, f := range f2 {

		find := false

		for _, e := range f1 {
			if f.Uin == e.Uin {
				find = true
				break
			}
		}

		if find {
			continue
		}

		f1 = append(f1, f)
	}

	total := len(f1)

	// if total <= cnt{
	// 	friends = f1
	// 	return
	// }

	a := rand.Perm(total)

	for _, idx := range a {
		friends = append(friends, f1[idx])

		if len(friends) >= cnt {
			break
		}
	}

	/*
		//如果不满足 则从未注册好友列表中选取邀请对象
		if len(friends) < cnt {

			_, t, err1 := GetRandomRecommendsFromAddrBookNotRegisted(uin, uuid, cnt-len(friends))
			if err1 != nil {
				log.Error(err1.Error())
				return
			}

			for _, v := range t {
				friends = append(friends, v)
			}
		}
	*/

	return
}

func GetRandomRecommendsFromSameSchoolExceptInvited(uin int64, inviteUins []int64, cnt int) (total int, friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

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

	strs := ""
	//排除好友
	for _, uid := range friendUins {
		strs += fmt.Sprintf("%d,", uid)
	}

	//排除已经邀请
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

	sql := fmt.Sprintf(`select count(uin) from profiles where schoolId = %d and uin not in (%s) and length(nickName) > 0`, ui.SchoolId, strs)

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

	if total < cnt {
		cnt = total
	}

	s := rand.Intn(total - cnt + 1)
	e := cnt

	sql = fmt.Sprintf(`select uin, phone, nickName, headImgUrl, gender, grade, schoolId, schoolType, schoolName from profiles where schoolId = %d and uin not in (%s) and length(nickName) > 0 limit %d, %d`, ui.SchoolId, strs, s, e)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	recommUins := make([]int64, 0)

	for rows.Next() {

		var fi RecommendInfo
		rows.Scan(&fi.Uin, &fi.Phone, &fi.NickName, &fi.HeadImgUrl, &fi.Gender, &fi.Grade, &fi.SchoolId, &fi.SchoolType, &fi.SchoolName)

		if len(fi.HeadImgUrl) > 0 {
			fi.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", fi.HeadImgUrl)
		}

		//初始为非好友状态
		fi.Status = constant.ENUM_SNS_STATUS_NOT_FRIEND
		fi.RecommendType = constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL
		fi.RecommendDesc = fmt.Sprintf("同校%s", st.GetGradeDescBySchool(fi.SchoolType, fi.Grade))

		friends = append(friends, &fi)
		recommUins = append(recommUins, fi.Uin)
	}

	/*
		    //是否已经发送加好友请求
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

func GetRandomRecommendsFromAddrBookHasRegistedExceptInvited(uin int64, uuid int64, inviteUins []int64, cnt int) (total int, friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

	//获取我的好友UIN
	friendUins, err := st.GetMyFriendUins(uin)
	if err != nil {
		return
	}

	strs := ""
	//排除好友
	for _, uid := range friendUins {
		strs += fmt.Sprintf("%d,", uid)
	}

	//排除已经邀请
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

	if total < cnt {
		cnt = total
	}

	s := rand.Intn(total - cnt + 1)
	e := cnt

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

func GetRandomRecommendsFromAddrBookNotRegisted(uin int64, uuid int64, cnt int) (total int, friends []*RecommendInfo, err error) {

	friends = make([]*RecommendInfo, 0)

	if uin == 0 {
		return
	}

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

	if total < cnt {
		cnt = total
	}

	s := rand.Intn(total - cnt + 1)
	e := cnt

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
