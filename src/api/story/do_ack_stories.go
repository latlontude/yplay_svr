package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
)

type AckStoriesReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	MinTs int64 `schema:"minTs"`
	MaxTs int64 `schema:"maxTs"`
}

type AckStoriesRsp struct {
	Cnt int `json:"cnt"`
}

func doAckStories(req *AckStoriesReq, r *http.Request) (rsp *AckStoriesRsp, err error) {

	log.Debugf("uin %d, AckStoriesReq %+v", req.Uin, req)

	cnt, err := AckStories(req.Uin, req.MinTs, req.MaxTs)
	if err != nil {
		log.Errorf("uin %d, AckStoriesRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AckStoriesRsp{cnt}

	log.Debugf("uin %d, AckStoriesRsp succ, %+v", req.Uin, rsp)

	return
}

func AckStories(uin int64, minTs, maxTs int64) (cnt int, err error) {

	if uin <= 0 || maxTs <= 0 || minTs <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	if maxTs < minTs {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "maxTs < minTs")
		log.Error(err.Error())
		return
	}

	//从redis获取最新一次拉取的时间
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//获取新的投票记录ID
	cnt, err = app.ZRemRangeByScore(keyStr, minTs, maxTs)
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
