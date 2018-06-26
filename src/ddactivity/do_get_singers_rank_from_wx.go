package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"common/util"
	"fmt"
	"net/http"
	"svr/st"
)

type GetSingerWithVoteFromWxReq struct {
	openId string `schema:"json:"openId"`
}

type GetSingerWithVoteFromWxRsp struct {
	SingersWithVote []SingerWithVoteInfo `json:"singersWithVote"`
	VotedSingerId   int                  `json:"votedSingerId"`
}

func doGetSingersRankingListFromWx(req *GetSingerWithVoteFromWxReq, r *http.Request) (rsp *GetSingerWithVoteFromWxRsp, err error) {
	log.Debugf("start doGetSingersRankingListFromWx openId:%s", req.openId)
	singersWithVote, err := GetSingersRankingListFromWx(req.openId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	singerId, err := GetMyVotedSingerIdFromWx(req.openId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerWithVoteFromWxRsp{singersWithVote, singerId}
	log.Debugf("end doGetSingersRankingListFromWx rsp:%+v", rsp)
	return
}

func GetSingersRankingListFromWx(openId string) (retSingersWithVotes []SingerWithVoteInfo, err error) {
	log.Debugf("start GetSingersRankingListFromPupu openId:%s", openId)

	singersWithVotes := make([]SingerWithVoteInfo, 0)
	retSingersWithVotes = make([]SingerWithVoteInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId, uin, rankActiveHeadImgUrl,  deptNamem, songName, songStoreName, songDuration from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singerWithVote SingerWithVoteInfo
		rows.Scan(&singerWithVote.SingerId, &singerWithVote.Uin, &singerWithVote.RankActiveHeadImgUrl, &singerWithVote.DeptName)

		singerWithVote.RankActiveHeadImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singerWithVote.RankActiveHeadImgUrl)

		ui, err1 := st.GetUserProfileInfo(singerWithVote.Uin)
		if err1 != nil {
			err = err1
			log.Errorf(err1.Error())
			return
		}

		singerWithVote.UserName = ui.UserName
		singerWithVote.NickName = ui.NickName
		singerWithVote.HeadImgUrl = ui.HeadImgUrl
		singerWithVote.Gender = ui.Gender
		singerWithVote.Grade = ui.Grade
		singersWithVotes = append(singersWithVotes, singerWithVote)
	}

	singerIdVoteMap, err := GetSingersVotedCnt()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for i, s := range singersWithVotes {
		if _, ok := singerIdVoteMap[s.SingerId]; ok {
			singersWithVotes[i].VoteCnt = singerIdVoteMap[s.SingerId]
		}
	}

	sortSingerIds := make([]int, 0)
	pairs := util.ReverseSortMap1(singerIdVoteMap)
	for _, p := range pairs {
		sortSingerIds = append(sortSingerIds, p.Key)
	}

	for _, singerId := range sortSingerIds {
		for i, s := range singersWithVotes {
			if s.SingerId == singerId {
				retSingersWithVotes = append(retSingersWithVotes, singersWithVotes[i])
				break
			}
		}
	}
	log.Debugf("end GetSingers retSingersWithVotes:%+v", retSingersWithVotes)

	return
}
