package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"svr/st"
)

type DelStoryReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"`
}

type DelStoryRsp struct {
}

func doDelStory(req *DelStoryReq, r *http.Request) (rsp *DelStoryRsp, err error) {

	log.Debugf("uin %d, DelStoryReq %+v", req.Uin, req)

	err = DelStory(req.Uin, req.StoryId)
	if err != nil {
		log.Errorf("uin %d, DelStoryRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelStoryRsp{}

	log.Debugf("uin %d, DelStoryRsp succ, %+v", req.Uin, rsp)

	return
}

func DelStory(uin, storyId int64) (err error) {
	log.Debugf("start DelStory uin:%d, storyId:%d", uin, storyId)

	if uin <= 0 || storyId <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	//从我的列表中删除
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_MY_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	_, err = app.ZRem(keyStr, fmt.Sprintf("%d_%d", uin, storyId))
	if err != nil {
		log.Error(err.Error())
		return
	}

	app, err = myredis.GetApp(constant.ENUM_REDIS_APP_STORY_MSG)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr = fmt.Sprintf("%d", storyId)

	err = app.Del(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	friendUins, err := st.GetMyFriendUins(uin)
	friendUins = append(friendUins, uin) //我也是我自己的好友
	go RemoveStoryByDelStory(uin, storyId, friendUins)
	log.Debugf("end DelStory")
	return
}

/*
解除好友关系，从我的新闻列表中移除好友发表的所有新闻
*/
func RemoveStoryByDelFriend(uin, uid int64) (err error) {

	log.Debugf("start RemoveStoryByDelFriend uin:%d, uid:%d", uin, uid)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)
	vals, err := app.ZRangeByScoreWithoutLimit(keyStr, -1, -1)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Debugf("vals:%+v", vals)

	removeStoryIds := make([]string, 0)

	for _, val := range vals {
		ret := strings.Split(val, "_")
		if len(ret) != 2 {
			log.Errorf("format err! val:%s", val)
			continue
		}

		storyUin, err1 := strconv.ParseInt(ret[0], 10, 64)
		if err1 != nil {
			log.Errorf("strconv.ParseInt err ret[0]:%s", ret[0])
		}

		if storyUin == uid {
			removeStoryIds = append(removeStoryIds, val)
		}
	}

	if len(removeStoryIds) > 0 {
		_, err1 := app.ZMRem(keyStr, removeStoryIds)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}
	}

	//用户A和B解除好友关系后，要从A发表的所有新闻中移除B的所有观看记录，并使这些观看记录总数减1
	go RemoveUserViewRecordByDelFriend(uin, uid)
	log.Debugf("end RemoveStoryByDelFriend uin:%d, uid:%d", uin, uid)
	return
}

/*
有用户删除自己发表的新闻，从该用户的所有好友的新闻列表移除该用户删除的新闻并移除该条新闻的观看记录和观看总数
*/
func RemoveStoryByDelStory(uin, storyId int64, friendsUins []int64) (err error) {

	log.Debugf("start RemoveStoryByDelStory uids:%+v, storyId:%d, uin:%d", friendsUins, storyId, uin)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, uid := range friendsUins {
		keyStr := fmt.Sprintf("%d", uid)
		_, err = app.ZRem(keyStr, fmt.Sprintf("%d_%d", uin, storyId))
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	//移除该条动态的观看记录和观看总数
	go RemoveStoryViewRecordByDelStory(uin, storyId)
	log.Debugf("end RemoveStoryByDelStory")
	return
}
