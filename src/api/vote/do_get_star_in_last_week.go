package vote

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type GetStarReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	User  int64  `schema:"uid"`
}

type GetStarRsp struct {
	Uin        int64  `schema:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	SchoolName string `json:"schoolName"`
	SchoolType int    `json:"schoolType"`
	Grade      int    `json:"grade"`
	Cnt        int    `json:"cnt"`        // 连续几周成为排行榜第一名
	DiamondSet []int  `josn:"diamondSet"` // 每周获得的钻石总数列表

	FriendUin        int64  `schema:"friendUin"`
	FriendNickName   string `json:"friendNickName"`
	FriendHeadImgUrl string `json:"friendHeadImgUrl"`
	FriendSchoolName string `json:"friendSchoolName"`
	FriendSchoolType int    `json:"friendSchoolType"`
	FriendGrade      int    `json:"friendGrade"`
	FriendCnt        int    `json:"friendCnt"`
	FriendDiamondSet []int  `json:"friendDiamondSet"`
}
type GetStarRspTmp struct {
	SameSchoolAndSameGradeWeekStar WeekStarUserInfo `json:"sameSchoolAndGradeWeekStar"`
	FriendsWeekStar                WeekStarUserInfo `json:"friendsWeekStar"`
}

type WeekStarUserInfo struct {
	Uin        int64  `schema:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	SchoolName string `json:"schoolName"`
	SchoolType int    `json:"schoolType"`
	Grade      int    `json:"grade"`
	Cnt        int    `json:"cnt"`        // 连续几周成为排行榜第一名
	DiamondSet []int  `josn:"diamondSet"` // 每周获得的钻石总数列表

}

func doGetStarInLastWeek(req *GetStarReq, r *http.Request) (rsp *GetStarRsp, err error) {

	log.Debugf("uin %d, GetStarInLastWeek %+v", req.Uin, req)

	info, err := GetStarOfWeek(req.User, 1) //查询上周话题明星
	if err != nil {
		log.Errorf("uin %d, GetStarInLastWeek error, %s", req.Uin, err.Error())
		return
	}

	last := 2 //查询上上周
	for {
		if info.SameSchoolAndSameGradeWeekStar.Uin == 0 && info.FriendsWeekStar.Uin == 0 {
			break
		}
		ret, _ := GetStarOfWeek(req.User, last)
		if ret.SameSchoolAndSameGradeWeekStar.Uin == info.SameSchoolAndSameGradeWeekStar.Uin && ret.SameSchoolAndSameGradeWeekStar.Uin != 0 {
			info.SameSchoolAndSameGradeWeekStar.Cnt++
			info.SameSchoolAndSameGradeWeekStar.DiamondSet = append(info.SameSchoolAndSameGradeWeekStar.DiamondSet, ret.SameSchoolAndSameGradeWeekStar.DiamondSet...)

		} else if ret.FriendsWeekStar.Uin == info.FriendsWeekStar.Uin && ret.FriendsWeekStar.Uin != 0 {
			info.FriendsWeekStar.Cnt++
			info.FriendsWeekStar.DiamondSet = append(info.FriendsWeekStar.DiamondSet, ret.FriendsWeekStar.DiamondSet...)

		} else {
			break
		}
		last++
	}

	var ret GetStarRsp
	ret.Uin = info.SameSchoolAndSameGradeWeekStar.Uin
	ret.NickName = info.SameSchoolAndSameGradeWeekStar.NickName
	ret.HeadImgUrl = info.SameSchoolAndSameGradeWeekStar.HeadImgUrl
	ret.SchoolName = info.SameSchoolAndSameGradeWeekStar.SchoolName
	ret.SchoolType = info.SameSchoolAndSameGradeWeekStar.SchoolType
	ret.Grade = info.SameSchoolAndSameGradeWeekStar.Grade
	ret.Cnt = info.SameSchoolAndSameGradeWeekStar.Cnt
	ret.DiamondSet = info.SameSchoolAndSameGradeWeekStar.DiamondSet

	ret.FriendUin = info.FriendsWeekStar.Uin
	ret.FriendNickName = info.FriendsWeekStar.NickName
	ret.FriendHeadImgUrl = info.FriendsWeekStar.HeadImgUrl
	ret.FriendSchoolName = info.FriendsWeekStar.SchoolName
	ret.FriendSchoolType = info.FriendsWeekStar.SchoolType
	ret.FriendGrade = info.FriendsWeekStar.Grade
	ret.FriendCnt = info.FriendsWeekStar.Cnt
	ret.FriendDiamondSet = info.FriendsWeekStar.DiamondSet

	rsp = &ret

	log.Debugf("uin %d, GetStarInLastWeek succ, %+v", req.Uin, rsp)

	return
}

func GetStarOfWeek(uin int64, last int) (ret GetStarRspTmp, err error) {

	log.Errorf("start GetStarOfWeek")
	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	e := GetFirstDayOfThisWeek()
	s := e - int64(last*7*24*3600)

	e = s + int64(7*24*3600)

	log.Errorf("start time:%s end time:%s", time.Unix(s, 0).Format("2006-01-02 15:04:05 PM"), time.Unix(e, 0).Format("2006-01-02 15:04:05 PM"))

	//要全部查询出来 然后找同校同年级的
	sql := fmt.Sprintf(`select voteToUin, count(id) as cnt from voteRecords where ts > %d and ts < %d group by voteToUin order by cnt desc`, s, e)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	in := false
	uidsSlice := make([]int64, 0)
	uidCntMap := make(map[int64]int)
	for rows.Next() {
		var uid int64
		var cnt int

		rows.Scan(&uid, &cnt)

		uidCntMap[uid] = cnt
		uidsSlice = append(uidsSlice, uid)

		if uid == uin {
			in = true
		}
	}

	if !in {
		uidsSlice = append(uidsSlice, uin) //也要获取本用户的信息，用于后续查找同校同年级的人
	}

	log.Errorf("uidsSlice:%+v", uidsSlice)

	res, err := st.BatchGetUserProfileInfo(uidsSlice)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	friendsUinsMap, err := GetAllMyFriends(uin)
	if err != nil {
		log.Errorf("failed to get all my friends")
		log.Errorf(err.Error())
		return
	}

	friendsUinsMap[uin] = 1 // 我也是我自己的好友

	var tmpRet GetStarRspTmp
	sameSchoolAndSameGradeWeekStarFlag := false
	friendsWeekStarFlag := false

	for _, uid := range uidsSlice {
		if _, ok := res[uid]; ok {
			if res[uin].SchoolId == res[uid].SchoolId && res[uin].Grade == res[uid].Grade && !sameSchoolAndSameGradeWeekStarFlag {

				if res[uin].SchoolType == 3 && res[uin].DeptId != res[uid].DeptId {
					continue // 用户学校为大学时，查找同校同学院同年级的用户
				}

				if uid != uin || (uid == uin && in) {
					tmpRet.SameSchoolAndSameGradeWeekStar.Uin = uid
					tmpRet.SameSchoolAndSameGradeWeekStar.NickName = res[uid].NickName
					tmpRet.SameSchoolAndSameGradeWeekStar.HeadImgUrl = res[uid].HeadImgUrl
					tmpRet.SameSchoolAndSameGradeWeekStar.SchoolName = res[uid].SchoolName
					tmpRet.SameSchoolAndSameGradeWeekStar.SchoolType = res[uid].SchoolType
					tmpRet.SameSchoolAndSameGradeWeekStar.Grade = res[uid].Grade
					tmpRet.SameSchoolAndSameGradeWeekStar.Cnt = 1
					tmpRet.SameSchoolAndSameGradeWeekStar.DiamondSet = append(tmpRet.SameSchoolAndSameGradeWeekStar.DiamondSet, uidCntMap[uid])
					sameSchoolAndSameGradeWeekStarFlag = true
				}
			}

			if _, ok := friendsUinsMap[uid]; ok && !friendsWeekStarFlag {
				if uid == uin && !in {
					continue
				}
				tmpRet.FriendsWeekStar.Uin = uid
				tmpRet.FriendsWeekStar.NickName = res[uid].NickName
				tmpRet.FriendsWeekStar.HeadImgUrl = res[uid].HeadImgUrl
				tmpRet.FriendsWeekStar.SchoolName = res[uid].SchoolName
				tmpRet.FriendsWeekStar.SchoolType = res[uid].SchoolType
				tmpRet.FriendsWeekStar.Grade = res[uid].Grade
				tmpRet.FriendsWeekStar.Cnt = 1
				tmpRet.FriendsWeekStar.DiamondSet = append(tmpRet.FriendsWeekStar.DiamondSet, uidCntMap[uid])
				friendsWeekStarFlag = true
			}

			if sameSchoolAndSameGradeWeekStarFlag && friendsWeekStarFlag {
				break
			}

		}
	}

	log.Errorf("star:%+v", tmpRet)
	ret = tmpRet
	log.Errorf("end GetStarOfWeek")
	return
}

func GetFirstDayOfThisWeek() (ts int64) {

	t := time.Now()
	week := int(t.Weekday())
	h, m, s := t.Clock()

	ts = t.Unix()

	ts -= int64((week-1)*24*3600 + h*3600 + m*60 + s)
	return

}
