package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	//"strconv"
)

type UpdateStoryViewRecordReq struct {
	Uin     int64  `schema:"uin"` // 观看此条动态的用户
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"` //观看的动态id
	Ts      int64  `schema:"ts"`      //  观看动态的时间
}

type UpdateStoryViewRecordRsp struct {
}

func doUpdateStoryViewRecord(req *UpdateStoryViewRecordReq, r *http.Request) (rsp *UpdateStoryViewRecordRsp, err error) {

	log.Debugf("uin %d, UpdateStoryViewRecordReq %+v", req.Uin, req)

	err = UpdateStoryViewRecord(req.Uin, req.StoryId, req.Ts)
	if err != nil {
		log.Errorf("uin %d, UpdateStoryViewRecordRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &UpdateStoryViewRecordRsp{}

	log.Debugf("uin %d, UpdateStoryViewRecordRsp succ, %+v", req.Uin, rsp)

	return
}

func UpdateStoryViewRecord(uin, storyId, ts int64) (err error) {

	log.Debugf("start UpdateStoryViewRecord")

	if uin <= 0 || storyId <= 0 || ts <= 0 {
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

	if !exist {
		score := ts // 该动态第一次被观看
		err = app.ZAdd(keyStr, score, member)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		log.Debugf("first view")
		err1 := app.Expire(keyStr, 86400) // 24小时过期
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}
		return
	}

	score := ts
	err = app.ZAdd(keyStr, score, member)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("end UpdateStoryViewRecord")
	return
}
