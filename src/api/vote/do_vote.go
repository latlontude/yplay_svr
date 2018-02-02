package vote

import (
	"api/im"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	//"strconv"
	"svr/st"
	"time"
)

type DoVoteReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	QId       int    `schema:"qid"`
	VoteToUin int64  `schema:"voteToUin"`
	Options   string `schema:"options"`
	Index     int    `schema:"index"` //题目编号
}

type DoVoteRsp struct {
	VoteRecordId int64 `json:"voteRecordId"`
}

func doVote(req *DoVoteReq, r *http.Request) (rsp *DoVoteRsp, err error) {

	log.Debugf("uin %d, DoVoteReq %+v", req.Uin, req)

	voteRecordId, err := Vote(req.Uin, req.QId, req.VoteToUin, req.Options, req.Index)
	if err != nil {
		log.Errorf("uin %d, DoVoteRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DoVoteRsp{voteRecordId}

	log.Debugf("uin %d, DoVoteRsp succ, %+v", req.Uin, rsp)

	return
}

func Vote(uin int64, qid int, voteToUin int64, optionStr string, index int) (voteRecordId int64, err error) {

	if uin == 0 || qid == 0 || len(optionStr) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	//非好友 直接忽略
	isFriend, err := st.IsFriend(uin, voteToUin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//无账号的不产生feed 不进行钻石计数
	//如果不是好友 也不产生feed 不进行钻石计数
	if voteToUin == 0 || isFriend == 0 {
		//添加到我的已投问题列表中
		//go Add2MyVotedQids(uin, qid)

		//服务器保存用户回答的进度,更新当前问题列表的已经回答的部分
		//st.UpdateVoteProgress2(uin, qid, index)

		st.UpdateVoteProgressByPreGene(uin, qid, index) //新接口

		log.Errorf("uin %d, voteToUin %d, qid %d, are not friend or voteToUin is 0", uin, voteToUin, qid)

		return
	}

	if uin == voteToUin {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "vote to self")
		log.Error(err.Error())
		return
	}

	//校验options信息是否正确
	var options []st.OptionInfo2
	err = json.Unmarshal([]byte(optionStr), &options)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, err.Error())
		log.Error(err.Error())
		return
	}

	if len(options) != constant.ENUM_OPTION_BATCH_SIZE {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid options size")
		log.Error(err.Error())
		return
	}

	go UserActRecords(uin, qid, 1)

	ts := time.Now().Unix()

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	status := constant.ENUM_VOTE_STATUS_INIT

	//IM会话ID
	imSessionId := ""

	stmt, err := inst.Prepare(`insert into voteRecords values(?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(0, uin, qid, voteToUin, optionStr, status, ts, imSessionId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	voteRecordId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//强制同步等待
	//if uin == 100328 || uin == 100446 {
	if true {
		err = st.UpdateVoteProgressByPreGene(uin, qid, index) //新接口
	} else {
		err = st.UpdateVoteProgress2(uin, qid, index) //新接口
	}
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//后续事项:
	//写入feeds表
	go GeneFeeds2(uin, qid, voteToUin, voteRecordId)

	//添加到我的已投问题列表中
	//go Add2MyVotedQids(uin, qid)

	//更新userStat中的钻石统计
	go st.IncrGemCnt(voteToUin)

	//更新userStat中的钻石统计
	go st.IncrUserQidVotedCnt(voteToUin, qid)

	//IM创建会话 然后发送第一条消息
	go im.SendVoteMsg(uin, qid, voteToUin, optionStr, voteRecordId)

	//被投票者生成消息 全部放在客户端的IM上做

	// 检查该题是否为投稿题，投稿人和被投人是否同校同年级
	uid, tpe, flag, _ := checkQidTypeAndSameSchoolSameGradeFlag(qid, voteToUin)
	if tpe == 1 && flag == 1 {
		// 通知用户，同校同年级的同学收到他(她)投稿的题的投票
		go im.SendSubmitVotedNotifyMsg(uid)
	}

	return
}

func checkQidTypeAndSameSchoolSameGradeFlag(qid int, votedUin int64) (submitUin int64, tpe int, flag int, err error) {

	log.Errorf("start checkQidTypeAndSameSchoolSameGradeFlag")

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select uin from submitQuestions where qid = %d and status = 1`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&submitUin)
	}

	if submitUin == 0 { // 不是投稿题目
		return
	}

	//是投稿题目
	tpe = 1

	uinsSlice := make([]int64, 0)
	uinsSlice = append(uinsSlice, submitUin)
	uinsSlice = append(uinsSlice, votedUin)

	res, err := st.BatchGetUserProfileInfo(uinsSlice)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if res[submitUin].SchoolId == res[votedUin].SchoolId && res[submitUin].Grade == res[votedUin].Grade {
		flag = 1
	}

	log.Errorf("end checkQidTypeAndSameSchoolSameGradeFlag")
	return
}

func UserActRecords(uin int64, qid int, act int) (err error) {

	ts := time.Now().Unix()
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into actRecords values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, qid, act, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return

}
func GeneFeeds2(uin int64, qid int, voteToUin int64, voteRecordId int64) (err error) {

	if uin == 0 || qid == 0 || voteToUin == 0 {
		return
	}

	friendUins, err := st.GetMyFriendUins(voteToUin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FEED_MSG)
	if err != nil {
		log.Error(err.Error())
		return
	}

	users := make([]int64, 0)

	for _, friendUin := range friendUins {

		//不用给自己发feed
		/*
			if friendUin == uin {
				continue
			}
		*/

		//uin 投票给voteToUin
		//那 voteToUin的好友都会收到一条feed
		ts := time.Now().UnixNano() / 1000000 //毫秒

		//friendUin的feed里面有一条feed表示 好友votToUin被uin投票选中了
		keyStr := fmt.Sprintf("%d", friendUin)
		err1 := app.ZAdd(keyStr, ts, fmt.Sprintf("%d", voteRecordId))
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		total, _ := app.ZCard(keyStr)

		//TrimCnt > MaxCnt 比如到600/500 到600的时候trim一次
		if total > env.Config.Feed.TrimCnt {
			log.Errorf("uin %d, trim feed msg, total %d", uin, total)

			_, err1 = app.ZRemRangeByRank(keyStr, 0, total-env.Config.Feed.MaxCnt-1)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}
		}

		users = append(users, friendUin)
	}

	//我的好友都会有新动态
	GeneNewFeedPush(users)

	return
}

func GeneNewFeedPush(uins []int64) (err error) {

	if len(uins) == 0 {
		return
	}

	//往channel里面放有新动态的用户
	for _, uin := range uins {
		im.ChanFeedPush <- uin
	}

	return
}
