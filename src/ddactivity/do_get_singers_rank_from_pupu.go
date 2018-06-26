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

type SingerWithVoteInfo struct {
	SingerId               int    `json:"singerId"`
	Uin                    int64  `json:"uin"`
	UserName               string `json:"userName"`
	NickName               string `json:"nickName"`
	HeadImgUrl             string `json:"headImgUrl"`
	Gender                 int    `json:"gender"`
	DeptName               string `json:"deptName"`
	Grade                  int    `json:"grade"`
	VoteCnt                int    `json:"voteCnt"`
	RankActiveHeadImgUrl   string `json:"rankActiveHeadImgUrl"`
	SingerDetailInfoImgUrl string `json:"singerDetailInfoImgUrl"`
	SingerSongUrl          string `json:"songUrl"`
	SingerSongName         string `json:"songName"`
	SingerSongDuration     string `json:"songDuration"`
}

type GetSingerWithVoteFromPupuReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetSingerWithVoteFromPupuRsp struct {
	SingersWithVote []SingerWithVoteInfo `json:"singersWithVote"`
	VotedSingerId   int                  `json:"votedSingerId"`
}

func doGetSingersRankingListFromPupu(req *GetSingerWithVoteFromPupuReq, r *http.Request) (rsp *GetSingerWithVoteFromPupuRsp, err error) {
	log.Debugf("start doGetSingersRankingListFromPupu uin:%d", req.Uin)
	singersWithVote, err := GetSingersRankingListFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	singerId, err := GetMyVotedSingerIdFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerWithVoteFromPupuRsp{singersWithVote, singerId}
	log.Debugf("end GetSingerWithVoteFromPupuRsp rsp:%+v", rsp)
	return
}

func GetSingersRankingListFromPupu(uin int64) (retSingersWithVotes []SingerWithVoteInfo, err error) {
	log.Debugf("start GetSingersRankingListFromPupu uin:%d", uin)

	singersWithVotes := make([]SingerWithVoteInfo, 0)
	retSingersWithVotes = make([]SingerWithVoteInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId, uin, rankActiveHeadImgUrl, singerDetailInfoImgUrl, deptName, songName, songStoreName, songDuration from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singerWithVote SingerWithVoteInfo
		var songStoreName string
		rows.Scan(&singerWithVote.SingerId, &singerWithVote.Uin, &singerWithVote.RankActiveHeadImgUrl, &singerWithVote.SingerDetailInfoImgUrl, &singerWithVote.DeptName, &singerWithVote.SingerSongName, &songStoreName, &singerWithVote.SingerSongDuration)

		singerWithVote.RankActiveHeadImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singerWithVote.RankActiveHeadImgUrl)
		singerWithVote.SingerDetailInfoImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singerWithVote.SingerDetailInfoImgUrl)
		singerWithVote.SingerSongUrl = "https://ddactive.yeejay.com/music/" + songStoreName

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

	//获取所有歌手
	sql := fmt.Sprintf(`select singerId from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	singerIds := make([]int, 0)
	for rows.Next() {
		var singerId int
		rows.Scan(&singerId)
		singerIds = append(singerIds, singerId)
	}

	//pupu渠道投票
	sql = fmt.Sprintf(`select uin, singerId from ddsingerFansFromPupu where status = 0`)
	rows, err = inst.Query(sql)
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

	for _, singerId := range singerIds {
		if _, ok := singerIdVoteMap[singerId]; !ok {
			singerIdVoteMap[singerId] = 0
		}
	}

	log.Debugf("end GetSingersVotedCnt singerIdVoteMap:%+v", singerIdVoteMap)
	return
}
