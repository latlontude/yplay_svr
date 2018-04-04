package activity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type GetSingerWithVoteFromWxReq struct {
	openId string `schema:"json:"openId"`
}

type GetSingerWithVoteFromWxRsp struct {
	SingersWithVote []SingerWithVoteInfo `json:"singersWithVote"`
}

func doGetSingersRankingListFromWx(req *GetSingerWithVoteFromWxReq, r *http.Request) (rsp *GetSingerWithVoteFromWxRsp, err error) {
	log.Debugf("start doGetSingersRankingListFromWx openId:%s", req.openId)
	singersWithVote, err := GetSingersRankingListFromWx(req.openId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerWithVoteFromWxRsp{singersWithVote}
	log.Debugf("end doGetSingersRankingListFromWx")
	return
}

func GetSingersRankingListFromWx(openId string) (singersWithVotes []SingerWithVoteInfo, err error) {
	log.Debugf("start GetSingersRankingListFromPupu openId:%s", openId)

	singersWithVotes = make([]SingerWithVoteInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId, uin, nickName, headImgUrl, gender, deptName, grade from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singerWithVote SingerWithVoteInfo
		rows.Scan(&singerWithVote.SingerId, &singerWithVote.Uin, &singerWithVote.NickName, &singerWithVote.HeadImgUrl, &singerWithVote.Gender, &singerWithVote.DeptName, &singerWithVote.Grade)
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

	log.Debugf("end GetSingers singersWithVotes:%+v", singersWithVotes)
	return
}
