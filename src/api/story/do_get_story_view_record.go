package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
)

type GetStoryViewRecordReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"`
}

type GetStoryViewRecordRsp struct {
	ViewInfo map[int64]int64 `json:"viewInfo"`
}

func doGetStoryViewRecord(req *GetStoryViewRecordReq, r *http.Request) (rsp *GetStoryViewRecordRsp, err error) {

	log.Debugf("uin %d, GetStoryViewRecordReq %+v", req.Uin, req)

	ret, err := GetStoryViewRecord(req.Uin, req.StoryId)
	if err != nil {
		log.Errorf("uin %d, GetStoryViewRecordRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetStoryViewRecordRsp{ret}

	log.Debugf("uin %d, GetStoryViewRecordRsp succ, %+v", req.Uin, rsp)

	return
}

func GetStoryViewRecord(uin, storyId int64) (ret map[int64]int64, err error) {
	log.Debugf("start GetStoryViewRecord uin:%d, storyId:%d", uin, storyId)

	ret = make(map[int64]int64)

	if storyId <= 0 {
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
	vals, err := app.ZRevRangeWithScores(keyStr, 0, -1)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(vals) == 0 {
		log.Debugf("nobody scan this msg")
		return
	}

	log.Debugf("valsStr:%+v", vals)

	if err != nil {
		log.Error(err.Error())
		return
	}

	var viewUid int64
	var viewTs int64

	for i, val := range vals {

		if i%2 == 0 {
			viewUid, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

		} else {

			viewTs, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				err = rest.NewAPIError(constant.E_REDIS_ZSET, "ZRevRangeWithScore value not interge")
				log.Error(err.Error())
				return
			}

			if viewTs > 0 && viewUid > 0 {
				ret[viewUid] = viewTs
			}
		}
	}

	log.Debugf("end GetStoryViewRecord ret:%+v", ret)
	return
}
