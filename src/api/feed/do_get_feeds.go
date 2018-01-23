package feed

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
	"svr/cache"
	"svr/st"
)

type GetFeedsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Ts  int64 `schema:"ts"` //比TS小的
	Cnt int   `schema:"cnt"`
}

type GetFeedsRsp struct {
	Feeds []*st.FeedInfo `json:"feeds"`
}

func doGetFeeds(req *GetFeedsReq, r *http.Request) (rsp *GetFeedsRsp, err error) {

	log.Debugf("uin %d, GetFeedsReq %+v", req.Uin, req)

	feeds, err := GetFeeds2(req.Uin, req.Ts, req.Cnt)
	if err != nil {
		log.Errorf("uin %d, GetFeedsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetFeedsRsp{feeds}

	log.Debugf("uin %d, GetFeedsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetFeeds2(uin int64, ts int64, cnt int) (feeds []*st.FeedInfo, err error) {

	feeds = make([]*st.FeedInfo, 0)

	if uin <= 0 || cnt <= 0 || ts <= 1000 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	//保存最近一次拉取feed的时间
	go UpdateFeedLastReadTs(uin, ts)

	//从redis获取最新一次拉取的时间
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FEED_MSG)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//获取新的投票记录ID
	valsStr, err := app.ZRevRangeByScoreWithScores(keyStr, ts-1, -1, 0, cnt)
	if err != nil {

		//如果KEY不存在,feed则为空
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				err = nil
				return
			}
		}

		log.Error(err.Error())
		return
	}

	//没有最新feed
	if len(valsStr) == 0 {
		return
	}

	if len(valsStr)%2 != 0 {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore return values cnt not even(2X)")
		log.Error(err.Error())
		return
	}

	var lastVid int64
	var lastMs int64

	vids := make(map[int64]int64)
	strs := "" //voteRecord拼接成数组供查询

	orderVids := make([]int64, 0)

	for i, valStr := range valsStr {

		if i%2 == 0 {
			lastVid, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

			strs += valStr + ","

		} else {

			lastMs, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

			if lastMs > 0 && lastVid > 0 {
				vids[lastVid] = lastMs

				orderVids = append(orderVids, lastVid)
			}
		}
	}

	if len(vids) == 0 {
		return
	}

	//去掉结尾的逗号
	strs = strs[:len(strs)-1]

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select id, uin, qid, voteToUin, ts from voteRecords where id in (%s)`, strs)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	//获取需要拉取用户资料的UIS列表
	uinsM := make(map[int64]int)
	feedsM := make(map[int64]*st.FeedInfo)

	for rows.Next() {
		var feed st.FeedInfo

		rows.Scan(&feed.VoteRecordId, &feed.VoteFromUin, &feed.QId, &feed.FriendUin, &feed.Ts)

		feed.Ts = vids[feed.VoteRecordId] //以实际写入消息队列的时间为准 单位为毫秒

		if q, ok := cache.QUESTIONS[feed.QId]; ok {
			feed.QText = q.QText
			feed.QIconUrl = q.QIconUrl
		}

		feedsM[feed.VoteRecordId] = &feed

		uinsM[feed.FriendUin] = 1
		uinsM[feed.VoteFromUin] = 1
	}

	//没有查询到
	if len(feedsM) == 0 || len(uinsM) == 0 {
		return
	}

	uinsA := make([]int64, 0)
	for uid, _ := range uinsM {
		uinsA = append(uinsA, uid)
	}

	//批量拉取用户资料
	res, err := st.BatchGetUserProfileInfo(uinsA)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//将拉取到的用户资料填充到返回结果中
	for _, vid := range orderVids {

		feed, ok := feedsM[vid]
		if !ok {
			continue
		}

		friendUin := feed.FriendUin
		voteFromUin := feed.VoteFromUin

		if info, ok := res[friendUin]; ok {

			feed.FriendNickName = info.NickName
			feed.FriendGender = info.Gender
			feed.FriendHeadImgUrl = info.HeadImgUrl
		}

		if info, ok := res[voteFromUin]; ok {

			feed.VoteFromGender = info.Gender
			feed.VoteFromSchoolId = info.SchoolId
			feed.VoteFromSchoolType = info.SchoolType
			feed.VoteFromSchoolName = info.SchoolName
			feed.VoteFromGrade = info.Grade

		}

		feeds = append(feeds, feed)
	}

	//隐藏投票者ID信息
	for i, _ := range feeds {
		feeds[i].VoteFromUin = 0
	}

	return
}

func UpdateFeedLastReadTs(uin int64, ts int64) {

	if uin == 0 || ts == 0 {
		return
	}

	//从redis获取最新一次拉取的时间
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_READ_FEED_MS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//获取新的投票记录ID
	valStr, err := app.Get(keyStr)
	if err != nil {

		//如果KEY不存在,feed则为空
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				err = nil
				valStr = "0"
			} else {
				log.Error(err.Error())
				return
			}
		} else {

			log.Error(err.Error())
			return
		}
	}

	lastMs, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if lastMs >= ts {
		return
	}

	err = app.Set(keyStr, fmt.Sprintf("%d", ts))
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}

func GetNewFeedsCnt(uin int64) (cnt int, err error) {

	//从redis获取最新一次拉取的时间
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_READ_FEED_MS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//获取新的投票记录ID
	valStr, err := app.Get(keyStr)
	if err != nil {

		//如果KEY不存在,feed则为空
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				err = nil
				valStr = "0"
			} else {
				log.Error(err.Error())
				return
			}
		} else {

			log.Errorf(err.Error())
			return
		}
	}

	lastMs, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		log.Error(err.Error())
		return
	}

	app, err = myredis.GetApp(constant.ENUM_REDIS_APP_FEED_MSG)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	cnt, err = app.ZCount(keyStr, lastMs, -1)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
