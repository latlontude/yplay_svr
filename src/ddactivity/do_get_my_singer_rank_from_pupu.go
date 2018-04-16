package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"common/util"
	"fmt"
	"net/http"
)

type GetMySingerRankFromPupuReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetMySingerRankFromPupuRsp struct {
	Status        int `json:"status"` // 0 其他错误，1 成功，2 还没有投票给任何人
	VotedSingerId int `json:"votedSingerId"`
	RankingNum    int `json:"rankingNum"`
}

func doGetMySingerRankFromPupu(req *GetMySingerRankFromPupuReq, r *http.Request) (rsp *GetMySingerRankFromPupuRsp, err error) {
	log.Debugf("start doGetMySingerRankFromPupu uin:%d", req.Uin)
	ret, err := GetMySingerRankFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &ret
	log.Debugf("end GetMySingerRankFromPupuRsp rsp:%+v", rsp)
	return
}

func GetMySingerRankFromPupu(uin int64) (ret GetMySingerRankFromPupuRsp, err error) {
	log.Debugf("start GetMySingerRankFromPupu uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId from ddsingerFansFromPupu where status = 0 and uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&ret.VotedSingerId)
	}

	if ret.VotedSingerId == 0 {
		ret.Status = 2
		log.Errorf("has not been any singers fans uin:%d", uin)
		return
	}

	singerIdVoteMap, err := GetSingersVotedCnt()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	singerIdRankMap := make(map[int]int)
	voteCnt := 0
	rankNum := 0
	pairs := util.ReverseSortMap1(singerIdVoteMap)
	for _, p := range pairs {
		if p.Value != voteCnt {
			voteCnt = p.Value
			rankNum++
		}
		singerIdRankMap[p.Key] = rankNum
	}

	ret.RankingNum = singerIdRankMap[ret.VotedSingerId]
	ret.Status = 1

	log.Debugf("end GetMySingerRankFromPupu:%+v", ret)
	return
}
