package vote

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"common/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"svr/st"
)

type GetRankingListReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	QId   int    `schema:"qid"`
}

type GetRankingListRsp struct {
	RankingInSameSchool        []UserInfo `json:"rankingInSameSchool"`
	RankingPercentInSameSchool string     `json:"rankingPercentInSameSchool"`
	RankingInFriends           []UserInfo `json:"rankingInFriends"`
	RankingPercentInFriends    string     `json:"rankingPercentInFriends"`
}

type UserInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	VotedCnt   int    `json:"votedCnt"`
}

func doGetRankingList(req *GetRankingListReq, r *http.Request) (rsp *GetRankingListRsp, err error) {

	log.Debugf("uin %d, GetRankingListReq %+v", req.Uin, req)

	info, err := GetRankingList(req.Uin, req.QId)
	if err != nil {
		log.Errorf("uin %d, GetRankingListRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &info

	log.Debugf("uin %d, GetRankigListRsp succ, %+v", req.Uin, rsp)

	return

}

func BatchGetRankingList(uin int64, qids []int) (retInfosMap map[int]GetRankingListRsp, err error) {
	log.Debugf("start BatchGetRankingList uin:%d", uin)
	var tmpRetInfo GetRankingListRsp
	retInfosMap = make(map[int]GetRankingListRsp)

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Errorf(err.Error())
		return
	}

	if len(qids) == 0 {
		return
	}

	str := ``
	for i, qid := range qids {

		if i != len(qids)-1 {
			str += fmt.Sprintf(`%d,`, qid)
		} else {
			str += fmt.Sprintf(`%d`, qid)
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	//要全部查询出来 然后找同校同年级的
	sql := fmt.Sprintf(`select qid, voteToUin, ts from voteRecords where qid in (%s)`, str)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	uidsSlice := make([]int64, 0)
	uidsMap := make(map[int64]int)

	votedQidsSlice := make([]int, 0)
	votedQidsMap := make(map[int]int)

	unvotedQidsSlice := make([]int, 0)

	qidsUinsCntMap := make(map[int]map[int64]int)
	qidsUinsMaxtsMap := make(map[int]map[int64]int)
	for rows.Next() {
		var qid int
		var uid int64
		var ts int
		rows.Scan(&qid, &uid, &ts)

		if _, ok := qidsUinsCntMap[qid]; !ok {
			qidsUinsCntMap[qid] = make(map[int64]int)
		}
		if _, ok := qidsUinsMaxtsMap[qid]; !ok {
			qidsUinsMaxtsMap[qid] = make(map[int64]int)
		}

		qidsUinsCntMap[qid][uid]++

		if _, ok := qidsUinsMaxtsMap[qid][uid]; ok {
			if ts > qidsUinsMaxtsMap[qid][uid] {
				qidsUinsMaxtsMap[qid][uid] = ts
			}
		} else {
			qidsUinsMaxtsMap[qid][uid] = ts
		}

		if _, ok := uidsMap[uid]; !ok {
			uidsSlice = append(uidsSlice, uid)
			uidsMap[uid] = 1
		}

		if _, ok := votedQidsMap[qid]; !ok {
			votedQidsSlice = append(votedQidsSlice, qid)
			votedQidsMap[qid] = 1
		}

	}

	for _, qid := range qids {
		if _, ok := votedQidsMap[qid]; !ok {
			unvotedQidsSlice = append(unvotedQidsSlice, qid)
		}
	}

	uidsSlice = append(uidsSlice, uin)
	res, err := st.BatchGetUserProfileInfo(uidsSlice)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	friendsUinsMap := make(map[int64]int)
	friendsUinsMap, err = GetAllMyFriends(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	friendsUinsMap[uin] = 1

	for _, qid := range votedQidsSlice {
		if _, ok := qidsUinsCntMap[qid]; ok {
			uidsCntMap := make(map[int]int)
			for uid, cnt := range qidsUinsCntMap[qid] {
				uidsCntMap[int(uid)] = cnt
			}

			userInfos := make([]UserInfo, 0)
			in := false
			//该题目下按用户被投数目最多的在最前
			pairs := util.ReverseSortMap1(uidsCntMap)
			for _, p := range pairs {
				var userInfo UserInfo
				userInfo.Uin = int64(p.Key)
				userInfo.VotedCnt = p.Value
				userInfos = append(userInfos, userInfo)

				if userInfo.Uin == uin {
					in = true
				}
			}

			//log.Errorf("before total:%d, userinfos:%+v", len(userInfos), userInfos)
			//对获得钻石数目相同的用户按照最新获得钻石的时间降序排列
			start := 0
			end := 0
			sameCnt := 0
			if len(userInfos) > 0 {
				sameCnt = userInfos[start].VotedCnt
			}

			for i := 1; i < len(userInfos); i++ {

				if userInfos[i].VotedCnt == sameCnt {
					end = i
				}

				if userInfos[i].VotedCnt != sameCnt || i == len(userInfos)-1 {
					if end != start {
						tmpUserInfos := userInfos[start : end+1]
						maxTsSlice := make([]int, 0)
						tsUinMap := make(map[int]int64)
						for _, info := range tmpUserInfos {
							maxTsSlice = append(maxTsSlice, qidsUinsMaxtsMap[qid][info.Uin])
							tsUinMap[qidsUinsMaxtsMap[qid][info.Uin]] = info.Uin
						}
						sort.Ints(maxTsSlice[:])
						for j := len(maxTsSlice) - 1; j >= 0; j-- {
							userInfos[start+(len(maxTsSlice)-1-j)].Uin = tsUinMap[maxTsSlice[j]]
						}

					}

					start = i
					end = start
					sameCnt = userInfos[start].VotedCnt
				}

			}
			//log.Errorf("after total:%d, userinfos:%+v", len(userInfos), userInfos)

			//该用户一定存在
			ui := res[uin]

			tmpUserInfos1 := make([]UserInfo, 0)
			allSameSchoolAndGradeCnt := 0
			myPos := 0
			flag := true

			for _, userInfo := range userInfos {

				if v, ok := res[userInfo.Uin]; ok {

					//同校同年级的前3名
					if v.SchoolId == ui.SchoolId && v.Grade == ui.Grade {
				/*		if v.SchoolType == 3 && v.DeptId != ui.DeptId {
							continue // 用户学校为大学时，查找同校同学院同年级的用户
						} */

						if flag {
							myPos++
						}

						if v.Uin == ui.Uin {
							flag = false
						}

						if len(tmpUserInfos1) < 3 {

							userInfo.NickName = v.NickName
							userInfo.HeadImgUrl = v.HeadImgUrl
							tmpUserInfos1 = append(tmpUserInfos1, userInfo)
						}
						allSameSchoolAndGradeCnt++
					}

				}
			}

			//	log.Errorf("allSameSchoolAndGradeCnt:%d", allSameSchoolAndGradeCnt)

			tmpRetInfo.RankingInSameSchool = tmpUserInfos1

			if allSameSchoolAndGradeCnt == 0 {
				tmpRetInfo.RankingPercentInSameSchool = "0%"
			} else {
				if allSameSchoolAndGradeCnt == 1 {
					if in {
						tmpRetInfo.RankingPercentInSameSchool = "100%"
					} else {
						tmpRetInfo.RankingPercentInSameSchool = "0%"
					}
				} else {
					tmpRetInfo.RankingPercentInSameSchool = strconv.Itoa(100*(allSameSchoolAndGradeCnt-myPos)/(allSameSchoolAndGradeCnt-1)) + "%"
				}
			}

			//	log.Errorf("qid:%d allSameSchoolAndGradeCnt:%d myPos in allSameSchoolAndGrade is:%d(%s) ", qid, allSameSchoolAndGradeCnt, myPos, tmpRetInfo.RankingPercentInSameSchool)

			tmpUserInfos2 := make([]UserInfo, 0)
			allVotedMyFriendsCnt := 0
			myPosInMyFriends := 0
			flag = true

			for _, userInfo := range userInfos {
				if _, ok := friendsUinsMap[userInfo.Uin]; ok {
					if v, ok := res[userInfo.Uin]; ok {

						if flag {
							myPosInMyFriends++
						}

						if v.Uin == ui.Uin {
							flag = false
						}

						if len(tmpUserInfos2) < 3 {

							userInfo.NickName = v.NickName
							userInfo.HeadImgUrl = v.HeadImgUrl

							tmpUserInfos2 = append(tmpUserInfos2, userInfo)
						}
						allVotedMyFriendsCnt++
					}

				}

			}

			//	log.Errorf("allVotedMyFriendsCnt:%d", allVotedMyFriendsCnt)

			tmpRetInfo.RankingInFriends = tmpUserInfos2
			if allVotedMyFriendsCnt == 0 {
				tmpRetInfo.RankingPercentInFriends = "0%"
			} else {
				if allVotedMyFriendsCnt == 1 {
					if in {
						tmpRetInfo.RankingPercentInFriends = "100%"
					} else {
						tmpRetInfo.RankingPercentInFriends = "0%"
					}
				} else {
					tmpRetInfo.RankingPercentInFriends = strconv.Itoa(100*(allVotedMyFriendsCnt-myPosInMyFriends)/(allVotedMyFriendsCnt-1)) + "%"
				}
			}
			//	log.Errorf("qid:%d allVotedMyFriendsCnt:%d myPos in allMyFriends is %d(%s)", qid, allVotedMyFriendsCnt, myPosInMyFriends, tmpRetInfo.RankingPercentInFriends)

			retInfosMap[qid] = tmpRetInfo
		}
	}

	for _, qid := range unvotedQidsSlice {
		var tmpRetInfo1 GetRankingListRsp
		retInfosMap[qid] = tmpRetInfo1
	}

	log.Debugf("end BatchGetRankingList uin:%d, retInfosMap:%+v", uin, retInfosMap)
	return
}

func GetRankingList(uin int64, qid int) (retInfo GetRankingListRsp, err error) {

	var tmpRetInfo GetRankingListRsp

	log.Errorf(" start GetRankingList")
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

	//要全部查询出来 然后找同校同年级的
	sql := fmt.Sprintf(`select voteToUin, count(id) as cnt from voteRecords where qid = %d group by voteToUin order by cnt desc`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	uins := make([]int64, 0)

	userInfos := make([]UserInfo, 0)
	in := false
	for rows.Next() {
		var userInfo UserInfo
		rows.Scan(&userInfo.Uin, &userInfo.VotedCnt)
		userInfos = append(userInfos, userInfo)
		uins = append(uins, userInfo.Uin)

		if userInfo.Uin == uin {
			in = true
		}

	}

	sql = fmt.Sprintf(`select voteToUin, ts from voteRecords where qid = %d order by ts desc`, qid)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	uinMaxTsMap := make(map[int64]int)
	for rows.Next() {
		var uid int64
		var ts int
		rows.Scan(&uid, &ts)
		if _, ok := uinMaxTsMap[uid]; ok {
			if ts > uinMaxTsMap[uid] {
				uinMaxTsMap[uid] = ts
			}
		} else {
			uinMaxTsMap[uid] = ts
		}
	}

	log.Errorf("before total:%d, userinfos:%+v", len(userInfos), userInfos)
	//对获得钻石数目相同的用户按照最新获得钻石的时间降序排列
	start := 0
	end := 0
	sameCnt := 0
	if len(userInfos) > 0 {
		sameCnt = userInfos[start].VotedCnt
	}

	for i := 1; i < len(userInfos); i++ {

		if userInfos[i].VotedCnt == sameCnt {
			end = i
		}

		if userInfos[i].VotedCnt != sameCnt || i == len(userInfos)-1 {
			if end != start {
				tmpUserInfos := userInfos[start : end+1]
				maxTsSlice := make([]int, 0)
				tsUinMap := make(map[int]int64)
				for _, info := range tmpUserInfos {
					maxTsSlice = append(maxTsSlice, uinMaxTsMap[info.Uin])
					tsUinMap[uinMaxTsMap[info.Uin]] = info.Uin
				}
				sort.Ints(maxTsSlice[:])
				for j := len(maxTsSlice) - 1; j >= 0; j-- {
					userInfos[start+(len(maxTsSlice)-1-j)].Uin = tsUinMap[maxTsSlice[j]]
				}

			}

			start = i
			end = start
			sameCnt = userInfos[start].VotedCnt
		}

	}
	log.Errorf("after total:%d, userinfos:%+v", len(userInfos), userInfos)

	if !in {
		uins = append(uins, uin)
	}

	res, err := st.BatchGetUserProfileInfo(uins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//该用户一定存在
	ui := res[uin]

	tmpUserInfos1 := make([]UserInfo, 0)
	allSameSchoolAndGradeCnt := 0
	myPos := 0
	flag := true

	for _, userInfo := range userInfos {

		if v, ok := res[userInfo.Uin]; ok {

			//同校同年级的前3名
			if v.SchoolId == ui.SchoolId && v.Grade == ui.Grade {
			/*	if v.SchoolType == 3 && v.DeptId != ui.DeptId {
					continue // 用户学校为大学时，查找同校同学院同年级的用户
				} */

				if flag {
					myPos++
				}

				if v.Uin == ui.Uin {
					flag = false
				}

				if len(tmpUserInfos1) < 3 {

					userInfo.NickName = v.NickName
					userInfo.HeadImgUrl = v.HeadImgUrl
					tmpUserInfos1 = append(tmpUserInfos1, userInfo)
				}
				allSameSchoolAndGradeCnt++
			}

		}
	}

	log.Errorf("allSameSchoolAndGradeCnt:%d", allSameSchoolAndGradeCnt)

	tmpRetInfo.RankingInSameSchool = tmpUserInfos1

	if allSameSchoolAndGradeCnt == 0 {
		tmpRetInfo.RankingPercentInSameSchool = "0%"
	} else {
		if allSameSchoolAndGradeCnt == 1 {
			if in {
				tmpRetInfo.RankingPercentInSameSchool = "100%"
			} else {
				tmpRetInfo.RankingPercentInSameSchool = "0%"
			}
		} else {
			tmpRetInfo.RankingPercentInSameSchool = strconv.Itoa(100*(allSameSchoolAndGradeCnt-myPos)/(allSameSchoolAndGradeCnt-1)) + "%"
		}
	}

	log.Errorf("qid:%d allSameSchoolAndGradeCnt:%d myPos in allSameSchoolAndGrade is:%d(%s) ", qid, allSameSchoolAndGradeCnt, myPos, tmpRetInfo.RankingPercentInSameSchool)

	friendsUinsMap := make(map[int64]int)
	friendsUinsMap, err = GetAllMyFriends(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	friendsUinsMap[uin] = 1

	tmpUserInfos2 := make([]UserInfo, 0)
	allVotedMyFriendsCnt := 0
	myPosInMyFriends := 0
	flag = true

	for _, userInfo := range userInfos {
		if _, ok := friendsUinsMap[userInfo.Uin]; ok {
			if v, ok := res[userInfo.Uin]; ok {

				if flag {
					myPosInMyFriends++
				}

				if v.Uin == ui.Uin {
					flag = false
				}

				if len(tmpUserInfos2) < 3 {

					userInfo.NickName = v.NickName
					userInfo.HeadImgUrl = v.HeadImgUrl

					tmpUserInfos2 = append(tmpUserInfos2, userInfo)
				}
				allVotedMyFriendsCnt++
			}

		}

	}

	log.Errorf("allVotedMyFriendsCnt:%d", allVotedMyFriendsCnt)

	tmpRetInfo.RankingInFriends = tmpUserInfos2
	if allVotedMyFriendsCnt == 0 {
		tmpRetInfo.RankingPercentInFriends = "0%"
	} else {
		if allVotedMyFriendsCnt == 1 {
			if in {
				tmpRetInfo.RankingPercentInFriends = "100%"
			} else {
				tmpRetInfo.RankingPercentInFriends = "0%"
			}
		} else {
			tmpRetInfo.RankingPercentInFriends = strconv.Itoa(100*(allVotedMyFriendsCnt-myPosInMyFriends)/(allVotedMyFriendsCnt-1)) + "%"
		}
	}
	log.Errorf("qid:%d allVotedMyFriendsCnt:%d myPos in allMyFriends is %d(%s)", qid, allVotedMyFriendsCnt, myPosInMyFriends, tmpRetInfo.RankingPercentInFriends)

	retInfo = tmpRetInfo
	log.Errorf("end GetRankingList")

	return

}

func GetAllMyFriends(uin int64) (friendsUinsMap map[int64]int, err error) {
	log.Errorf("start GetAllMyFriends")
	friendsUinsMap = make(map[int64]int)

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

	sql := fmt.Sprintf(`select friendUin from friends where uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		friendsUinsMap[uid] = 1
	}
	log.Errorf("end GetAllMyFriends")
	return

}
