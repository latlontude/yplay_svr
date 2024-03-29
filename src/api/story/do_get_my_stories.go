package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	RetStories []*st.RetStoryInfo `json:"stories"`
}

func doGetMyStories(req *GetMyStoriesReq, r *http.Request) (rsp *GetMyStoriesRsp, err error) {

	log.Debugf("uin %d, GetMyStoriesReq %+v", req.Uin, req)

	retStories, err := GetMyStories(req.Uin, req.Ts, req.Cnt)
	if err != nil {
		log.Errorf("uin %d, GetMyStoriesRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetMyStoriesRsp{retStories}

	log.Debugf("uin %d, GetMyStoriesRsp succ, %+v", req.Uin, rsp)

	return
}

func GetMyStories(uin int64, ts int64, cnt int) (retStories []*st.RetStoryInfo, err error) {
	log.Debugf("start GetMyStories uin:%d ts:%d cnt:%d", uin, ts, cnt)

	stories := make([]*st.StoryInfo, 0)
	retStories = make([]*st.RetStoryInfo, 0)

	if uin <= 0 || cnt <= 0 {
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

	if ts == 0 { // ts 为0 代表获取最新的
		ts = time.Now().UnixNano()
	}
	//获取新的STORY ID
	valsStr, err := app.ZRevRangeByScoreWithScores(keyStr, ts-1, -1, 0, cnt)
	if err != nil {

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

			ret := strings.Split(valStr, "_")
			if len(ret) != 2 {
				log.Debugf("valStr:%s does not fit uid_storyId", valStr)
				lastVid = 0
			} else {
				lastVid, err = strconv.ParseInt(ret[1], 10, 64)
				if err != nil {
					err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
					log.Error(err.Error())
					return
				}
			}

		} else {

			lastMs, err = strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

			if lastMs > 0 && lastVid > 0 && lastMs == lastVid {
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
	uinsM := make(map[int64]int)

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
			uinsM[si.Uin] = 1
		}
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

	for _, story := range stories {

		var retStory st.RetStoryInfo
		retStory.StoryId = story.StoryId
		retStory.Type = story.Type
		retStory.Text = story.Text
		retStory.Data = story.Data
		retStory.Uin = story.Uin
		retStory.ThumbnailImgUrl = story.ThumbnailImgUrl
		retStory.ViewCnt = story.ViewCnt
		retStory.Ts = story.Ts

		if _, ok := res[story.Uin]; ok {
			retStory.NickName = res[story.Uin].NickName
			retStory.HeadImgUrl = res[story.Uin].HeadImgUrl
		}

		retStories = append(retStories, &retStory)
	}
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
