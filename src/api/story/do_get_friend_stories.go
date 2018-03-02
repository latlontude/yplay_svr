package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"svr/st"
	"time"
)

type GetFriendStoriesReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Ts    int64  `schema:"ts"` //比TS小的
	Cnt   int    `schema:"cnt"`
}

type GetFriendStoriesRsp struct {
	Stories []*st.StoryInfo `json:"stories"`
}

func doGetFriendStories(req *GetFriendStoriesReq, r *http.Request) (rsp *GetFriendStoriesRsp, err error) {

	log.Debugf("uin %d, GetFriendStoriesReq %+v", req.Uin, req)

	stories, err := GetFriendStories(req.Uin, req.Ts, req.Cnt)
	if err != nil {
		log.Errorf("uin %d, GetFriendStoriesRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetFriendStoriesRsp{stories}

	log.Debugf("uin %d, GetFriendStoriesRsp succ, %+v", req.Uin, rsp)

	return
}

func GetFriendStories(uin int64, ts int64, cnt int) (stories []*st.StoryInfo, err error) {

	stories = make([]*st.StoryInfo, 0)

	if uin <= 0 || cnt <= 0 || ts <= 1000 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	//保存最近一次拉取feed的时间
	go UpdateStoryLastReadTs(uin)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//先删除24小时之前发表的
	expireTs := time.Now().UnixNano()/1000000 - 86400000
	_, err1 := app.ZRemRangeByScore(keyStr, 0, expireTs)
	if err1 != nil {
		log.Errorf(err1.Error())
	}

	//获取新的STORY ID
	valsStr, err := app.ZRevRangeByScoreWithScores(keyStr, ts-1, -1, 0, cnt)
	if err != nil {

		//如果KEY不存在,feed则为空
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				err = nil
				return
			}
		}

		log.Errorf(err.Error())
		return
	}

	//没有最新story
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

	orderVids := make([]int64, 0)

	for i, valStr := range valsStr {

		if i%2 == 0 {
			lastVid, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

		} else {

			lastMs, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

			if lastMs > 0 && lastVid > 0 {
				orderVids = append(orderVids, lastVid)
			}
		}
	}

	if len(orderVids) == 0 {
		return
	}

	app2, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_MSG)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	storyIds := make([]string, 0)
	for _, vid := range orderVids {
		storyIds = append(storyIds, fmt.Sprintf("%d", vid))
	}

	storyVals, err := app2.MGet(storyIds)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//获取需要拉取用户资料的UIS列表
	//uinsM := make(map[int64]int)

	//将拉取到的用户资料填充到返回结果中
	for _, svid := range storyIds {

		storyVal, ok := storyVals[svid]
		if !ok {
			continue
		}

		var si st.StoryInfo
		err = json.Unmarshal([]byte(storyVal), &si)
		if err != nil {
			log.Errorf(err.Error())
			return
		} else {
			storyId, _ := strconv.ParseInt(svid, 10, 64)
			viewRecord, _ := GetStoryViewRecord(storyId)
			si.ViewCnt = len(viewRecord)
			stories = append(stories, &si)
			//uinsM[si.Uin] = 1
		}
	}

	/*
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
	*/

	return
}

func UpdateStoryLastReadTs(uin int64) {

	if uin == 0 {
		return
	}

	//从redis获取最新一次拉取的时间
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_READ_STORY_MS)
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

	ms := time.Now().UnixNano() / 1000000

	if lastMs >= ms {
		return
	}

	err = app.Set(keyStr, fmt.Sprintf("%d", ms))
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}

func GetNewStoriesCnt(uin int64) (cnt int, err error) {

	//从redis获取最新一次拉取的时间
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_READ_STORY_MS)
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

	app, err = myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
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
