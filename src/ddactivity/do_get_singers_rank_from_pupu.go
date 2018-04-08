package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type SingerWithVoteInfo struct {
	SingerId   int    `json:"singerId"`
	Uin        int    `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	DeptName   string `json:"deptName"`
	Grade      int    `json:"grade"`
	VoteCnt    int    `json:"voteCnt"`
}

type GetSingerWithVoteFromPupuReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetSingerWithVoteFromPupuRsp struct {
	SingersWithVote []SingerWithVoteInfo `json:"singersWithVote"`
}

func doGetSingersRankingListFromPupu(req *GetSingerWithVoteFromPupuReq, r *http.Request) (rsp *GetSingerWithVoteFromPupuRsp, err error) {
	log.Debugf("start doGetSingersRankingListFromPupu uin:%d", req.Uin)
	singersWithVote, err := GetSingersRankingListFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerWithVoteFromPupuRsp{singersWithVote}
	log.Debugf("end GetSingerWithVoteFromPupuRsp")
	return
}

func GetSingersRankingListFromPupu(uin int64) (singersWithVotes []SingerWithVoteInfo, err error) {
	log.Debugf("start GetSingersRankingListFromPupu uin:%d", uin)

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

func GetSingersVotedCnt() (singerIdVoteMap map[int]int, err error) {
	log.Debugf("start GetSingersVotedCnt")
	singerIdVoteMap = make(map[int]int)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//pupu渠道投票
	sql := fmt.Sprintf(`select uin, singerId from ddsingerFansFromPupu where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singerId int
		var uid int64
		rows.Scan(&uid, &singerId)
		if singerId != 0 {
			singerIdVoteMap[singerId]++
		}
	}

	//微信渠道投票
	sql = fmt.Sprintf(`select openId, singerId from ddsingerFansFromWx where status = 0`)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singerId int
		var openId string
		rows.Scan(&openId, &singerId)
		if singerId != 0 {
			singerIdVoteMap[singerId]++
		}
	}

	log.Debugf("end GetSingersVotedCnt singerIdVoteMap:%+v", singerIdVoteMap)
	return
}
