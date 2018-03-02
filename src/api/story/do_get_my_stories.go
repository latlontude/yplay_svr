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

type GetMyStoriesReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Ts    int64  `schema:"ts"` //比TS小的
	Cnt   int    `schema:"cnt"`
}

type GetMyStoriesRsp struct {
	Stories []*st.StoryInfo `json:"stories"`
}

func doGetMyStories(req *GetMyStoriesReq, r *http.Request) (rsp *GetMyStoriesRsp, err error) {

	log.Debugf("uin %d, GetMyStoriesReq %+v", req.Uin, req)

	stories, err := GetMyStories(req.Uin, req.Ts, req.Cnt)
	if err != nil {
		log.Errorf("uin %d, GetMyStoriesRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetMyStoriesRsp{stories}

	log.Debugf("uin %d, GetMyStoriesRsp succ, %+v", req.Uin, rsp)

	return
}

func GetMyStories(uin int64, ts int64, cnt int) (stories []*st.StoryInfo, err error) {
	log.Debugf("start GetMyStories uin:%d", uin)

	stories = make([]*st.StoryInfo, 0)

	if uin <= 0 || cnt <= 0 || ts <= 1000 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_MY_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//获取24小时之前发表的
	expireTs := time.Now().UnixNano()/1000000 - 86400000
	vals, err := app.ZRangeByScoreWithoutLimit(keyStr, -1, expireTs)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(vals) > 0 { // 有24小时之前发表的动态

		log.Debugf("have %d subjects before 24 hours", len(vals))

		app1, err1 := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}

		//从动态观看总数列表列表中移除该用户24小时之前发表的动态
		keyStr1 := fmt.Sprintf("total")
		_, err1 = app1.ZMRem(keyStr1, vals)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}

	} else {
		log.Debugf("no subjects before 24 hours")
	}

	//先删除24小时之前发表的
	_, err = app.ZRemRangeByScore(keyStr, 0, expireTs)
	if err != nil {
		log.Errorf(err.Error())
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
		log.Debugf("have not new story")
		return
	}

	log.Debugf("valsStr:%+v", valsStr)

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

	log.Debugf("orderVids:%+v", orderVids)

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

	log.Debugf("storyVals:%+v", storyVals)

	// 获取动态观看总数
	app3, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	viewVals, err := app3.ZRangeWithScores("total", 0, -1)
	if err != nil {
		log.Errorf(err.Error())
		log.Errorf("failed to  get view cnt")
		return
	}

	log.Debugf("viewVals:%+v", viewVals)

	storyviewCntIdMap := make(map[string]int64)
	var storyId string
	var viewCnt int64

	for i, viewVal := range viewVals {
		if i%2 == 0 {
			storyId = viewVal
		} else {

			viewCnt, err = strconv.ParseInt(viewVal, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

			if len(storyId) > 0 {
				storyviewCntIdMap[storyId] = viewCnt
			}
		}
	}

	log.Debugf("storyviewCntIdMap %+v", storyviewCntIdMap)

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
			if cnt, ok := storyviewCntIdMap[svid]; ok {
				si.ViewCnt = int(cnt)
			}
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
	log.Debugf("end GetMyStories uin:%d", uin)
	return
}

func GetMyStoriesCnt(uin int64) (total int, err error) {
	log.Debugf("start GetMyStoriesCnt uin:%d", uin)

	if uin <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_MY_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)
	//先删除24小时之前发表的
	expireTs := time.Now().UnixNano()/1000000 - 86400000
	_, err = app.ZRemRangeByScore(keyStr, 0, expireTs)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	cnt, err := app.ZCard(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		log.Debugf("get stories count error")
		return
	}

	total = cnt
	log.Debugf("end GetMyStoriesCnt uin:%d", uin)
	return
}

func GetUserStories(uin, uid, ts int64, cnt int) (stories []*st.StoryInfo, err error) {

	log.Debugf("start GetUserStories uin:%d, uid:%d", uin, uid)
	stories = make([]*st.StoryInfo, 0)

	if uin <= 0 || uid <= 0 || ts <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_MY_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uid)

	//获取24小时之前发表的
	expireTs := time.Now().UnixNano()/1000000 - 86400000
	vals, err := app.ZRangeByScoreWithoutLimit(keyStr, -1, expireTs)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(vals) > 0 { // 有24小时之前发表的动态

		log.Debugf("have %d subjects before 24 hours", len(vals))

		app1, err1 := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}

		//从动态观看总数列表列表中移除该用户24小时之前发表的动态
		keyStr1 := fmt.Sprintf("total")
		_, err1 = app1.ZMRem(keyStr1, vals)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}

	} else {
		log.Debugf("no subjects before 24 hours")
	}

	//先删除24小时之前发表的
	_, err = app.ZRemRangeByScore(keyStr, 0, expireTs)
	if err != nil {
		log.Errorf(err.Error())
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
		log.Debugf("have not new story")
		return
	}

	log.Debugf("valsStr:%+v", valsStr)

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

	log.Debugf("orderVids:%+v", orderVids)

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

	log.Debugf("storyVals:%+v", storyVals)

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
			stories = append(stories, &si)

		}
	}
	log.Debugf("end GetUserStories")
	return
}
