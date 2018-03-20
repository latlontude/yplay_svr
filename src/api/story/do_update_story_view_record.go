package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type UpdateStoryViewRecordReq struct {
	Uin     int64  `schema:"uin"` // 观看此条动态的用户
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"` //观看的动态id
}

type UpdateStoryViewRecordRsp struct {
}

func doUpdateStoryViewRecord(req *UpdateStoryViewRecordReq, r *http.Request) (rsp *UpdateStoryViewRecordRsp, err error) {

	log.Debugf("uin %d, UpdateStoryViewRecordReq %+v", req.Uin, req)

	err = UpdateStoryViewRecord(req.Uin, req.StoryId)
	if err != nil {
		log.Errorf("uin %d, UpdateStoryViewRecordRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &UpdateStoryViewRecordRsp{}

	log.Debugf("uin %d, UpdateStoryViewRecordRsp succ, %+v", req.Uin, rsp)

	return
}

func UpdateStoryViewRecord(uin, storyId int64) (err error) {

	log.Debugf("start UpdateStoryViewRecord")

	if uin <= 0 || storyId <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", storyId)
	member := fmt.Sprintf("%d", uin)

	exist, err := app.Exist(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	viewTs := time.Now().UnixNano() / 1000000

	if !exist {

		//加入该条动态的观看者记录列表
		score1 := viewTs
		err = app.ZAdd(keyStr, score1, member)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		//更新该条动态的总观看人数
		keyStr1 := "total"
		member1 := fmt.Sprintf("%d", storyId)
		_, err = app.ZIncrBy(keyStr1, member1, 1)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		//该条动态首次被观看，设置观看者列表过期时间
		log.Debugf("first view")
		err1 := app.Expire(keyStr, 86400) // 24小时过期
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}

		return
	}

	orgCnt, err := app.ZCard(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	score := viewTs
	err = app.ZAdd(keyStr, score, member)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	nowCnt, err := app.ZCard(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if nowCnt == orgCnt+1 {
		//更新该条动态的总观看人数
		keyStr = "total"
		member = fmt.Sprintf("%d", storyId)
		_, err = app.ZIncrBy(keyStr, member, 1)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
	}

	log.Debugf("end UpdateStoryViewRecord")
	return
}

/*
用户A和B解除好友关系后，要从A发表的所有新闻中移除B的所有观看记录，并使这些观看记录总数减1
*/
func RemoveUserViewRecordByDelFriend(uin, uid int64) (err error) {

	log.Debugf("start RemoveUserViewRecord uin:%d, uid:%d", uin, uid)
	if uin <= 0 || uid <= 0 {
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

	app1, err1 := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
	if err1 != nil {
		log.Error(err1.Error())
		return
	}
	keyStr1 := fmt.Sprintf("total")

	if len(vals) > 0 { // 有24小时之前发表的动态
		log.Debugf("have %d subjects before 24 hours", len(vals))

		//从动态观看总数列表列表中移除该用户24小时之前发表的动态
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

	//获取所有storyId
	vals, err = app.ZRange(keyStr, 0, -1)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//没有最新的story
	if len(vals) == 0 {
		log.Debugf("have not new story")
		return
	}

	remViewRecordStoryIds := make([]string, 0) // 存储uid看过uin发表的所有新闻的storyId
	for _, storyIdStr := range vals {
		storyId, err1 := strconv.ParseInt(storyIdStr, 10, 64)
		if err1 != nil {
			log.Errorf("failed in strconv.ParseInt :%s", storyIdStr)
			continue
		}

		viewerInfos, err1 := GetStoryViewRecord(uin, storyId)
		if err1 != nil {
			log.Errorf("failed to get storyId:%s view record", storyIdStr)
			continue
		}

		for _, info := range viewerInfos {
			if info.Uin == uid {
				remViewRecordStoryIds = append(remViewRecordStoryIds, storyIdStr)
			}
		}
	}

	log.Debugf("remViewRecordStoryIds:%+v", remViewRecordStoryIds)

	member := fmt.Sprintf("%s", uid)
	for _, storyId := range remViewRecordStoryIds {
		_, err1 := app1.ZIncrBy(keyStr1, storyId, -1)
		if err1 != nil {
			log.Errorf("failed to down  storyId:%s view total", storyId)
			continue
		}

		_, err1 = app1.ZRem(storyId, member)
		if err1 != nil {
			log.Errorf("failed to remove storyId:%s view record", storyId)
			continue
		}
	}

	log.Debugf("end RemoveUserViewRecord")
	return
}

/*
 用户删除自己发表的动态后，要删除该条动态的观看记录和该条动态的总观看人数
*/
func RemoveStoryViewRecordByDelStory(uin, storyId int64) (err error) {
	log.Debugf("start RemoveStoryViewRecordByDelStory uin:%d storyId:%d", uin, storyId)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//移除观看记录
	keyStr := fmt.Sprintf("%d", storyId) // 不用移除也可以，24小时自动过期
	err = app.Del(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//移除这条记录的观看总数
	keyStr = "total"
	member := fmt.Sprintf("%d", storyId)
	_, err = app.ZRem(keyStr, member)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("end RemoveStoryViewRecordByDelStory")
	return
}
