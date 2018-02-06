package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
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

func DelStory(uin int64, storyId int64) (err error) {

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

	//获取新的投票记录ID
	_, err = app.ZRem(keyStr, fmt.Sprintf("%d", storyId))
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
	//获取新的投票记录ID
	err = app.Del(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
