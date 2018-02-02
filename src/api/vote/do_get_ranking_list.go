package vote

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
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
