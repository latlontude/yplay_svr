package story

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	//"strconv"
	"svr/st"
	"time"
)

type AddStoryReq struct {
	Uin             int64  `schema:"uin"`
	Token           string `schema:"token"`
	Ver             int    `schema:"ver"`
	Type            int    `schema:"type"`
	Data            string `schema:"data"`
	Text            string `schema:"text"`
	ThumbnailImgUlr string `schema:"thumbnailImgUlr"`
}

type AddStoryRsp struct {
	StoryId int64 `json:"storyId"`
}

func doAddStory(req *AddStoryReq, r *http.Request) (rsp *AddStoryRsp, err error) {

	log.Debugf("uin %d, AddStoryReq %+v", req.Uin, req)

	sid, err := AddStory(req.Uin, req.Type, req.Data, req.Text, req.ThumbnailImgUlr)
	if err != nil {
		log.Errorf("uin %d, AddStoryRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddStoryRsp{sid}

	log.Debugf("uin %d, AddStoryRsp succ, %+v", req.Uin, rsp)

	return
}

func AddStory(uin int64, typ int, data, text, thumbnailImgUrl string) (sid int64, err error) {

	sid = time.Now().UnixNano() / 1000000

	if uin <= 0 || typ <= 0 || len(text) == 0 || len(data) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	var si st.StoryInfo
	si.StoryId = sid
	si.Text = text
	si.Data = data
	si.Type = typ
	si.Ts = sid
	si.Uin = uin
	si.ThumbnailImgUrl = thumbnailImgUrl
	si.ViewCnt = 0

	d, err := json.Marshal(&si)
	if err != nil {
		log.Errorf("json marshal error", err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//存储story信息 并设置过期时间
	valStr := string(d)
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_MSG)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//24小时过期
	err = app.SetEx(keyStr, valStr, 86400)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//将story加入到我的列表里面
	app, err = myredis.GetApp(constant.ENUM_REDIS_APP_MY_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//先删除24小时之前发表的
	expireTs := time.Now().UnixNano()/1000000 - 86400000
	_, err1 := app.ZRemRangeByScore(keyStr, 0, expireTs)
	if err1 != nil {
		log.Errorf(err1.Error())
	}

	sidStr := fmt.Sprintf("%d", sid)

	err = app.ZAdd(keyStr, sid, sidStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//在我的好友列表中插入新story
	go GeneNewStory(uin, sid)

	return
}

func GeneNewStory(uin int64, storyId int64) (err error) {

	if uin == 0 || storyId == 0 {
		return
	}

	friendUins, err := st.GetMyFriendUins(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//users := make([]int64, 0)

	for _, friendUin := range friendUins {

		//friendUin的story里面有一条story表示 好友uin有新的story了
		keyStr := fmt.Sprintf("%d", friendUin)
		err1 := app.ZAdd(keyStr, storyId, fmt.Sprintf("%d", storyId))
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		//users = append(users, friendUin)
	}

	//我的好友都会有新story
	//GeneNewFeedPush(users)

	return
}
