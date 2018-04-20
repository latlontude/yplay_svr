package story

import (
	"api/im"
	"common/constant"
	"common/env"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type PictureData struct {
	ImgLargeUrl string `json:"imgLargeUrl"`
	ImgOrgUrl   string `json:"imgOrgUrl"`
}
type VideoData struct {
	VideoDuration int    `json:"videoDuration"`
	VideoUrl      string `json:"videoUrl"`
}

type AddStoryReq struct {
	Uin             int64  `schema:"uin"`
	Token           string `schema:"token"`
	Ver             int    `schema:"ver"`
	Type            int    `schema:"type"` // 1 文字+图片 2 文字+视频 3 文字+gif动图
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

	log.Debugf("start AddStory uin:%d, typ:%d, data:%s, text:%s,thumbnailImgUrl: %s", uin, typ, data, text, thumbnailImgUrl)

	sid = time.Now().UnixNano() / 1000000

	if uin <= 0 || (typ != 1 && typ != 2 && typ != 3) || len(text) == 0 || len(data) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	if typ == 1 || typ == 3 {
		var v PictureData
		err = json.Unmarshal([]byte(data), &v)
		if err != nil {
			log.Errorf("decode json err!")
			return
		}
	} else if typ == 2 {
		var v VideoData
		err = json.Unmarshal([]byte(data), &v)
		if err != nil {
			log.Errorf("decode json err!")
			return
		}
	}

	go RecordStory(uin, sid, typ, data, text, thumbnailImgUrl)

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

	keyStr := fmt.Sprintf("%d", sid)

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

	//获取该用户24前发表的所有动态
	keyStr = fmt.Sprintf("%d", uin)
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
	expireTs = time.Now().UnixNano()/1000000 - 86400000
	_, err = app.ZRemRangeByScore(keyStr, 0, expireTs)
	if err != nil {
		log.Errorf(err.Error())
	}

	sidStr := fmt.Sprintf("%d_%d", uin, sid) // member:uid_storyId

	err = app.ZAdd(keyStr, sid, sidStr) // score member
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//在我的好友列表中插入新story
	go GeneNewStory(uin, sid)

	log.Debugf("end AddStory")
	return
}

func GeneNewStory(uin, storyId int64) (err error) {
	log.Debugf("start GeneNewStory uin:%d, storyId:%d", uin, storyId)

	if uin == 0 || storyId == 0 {
		return
	}

	friendUins, err := st.GetMyFriendUins(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	friendUins = append(friendUins, uin) //朋友圈也能看到自己发表的动态
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_FRIEND_STORY_LIST)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, friendUin := range friendUins {

		keyStr := fmt.Sprintf("%d", friendUin)
		total, _ := app.ZCard(keyStr)
		if total > env.Config.Story.TrimCnt { //如果用户有很多好友，但该好友如果长时间不登陆，可能会导致用户的新闻列表越来越长
			log.Debugf("news total > %d , trim it", env.Config.Story.TrimCnt)

			//删除24小时之前发表的
			expireTs := time.Now().UnixNano()/1000000 - 86400000
			_, err = app.ZRemRangeByScore(keyStr, 0, expireTs)
			if err != nil {
				log.Errorf(err.Error())
				continue
			}
		}

		//friendUin的story里面有一条story表示 好友uin有新的story了
		err1 := app.ZAdd(keyStr, storyId, fmt.Sprintf("%d_%d", uin, storyId))
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}
	}

	//我的好友都会有新story
	GeneNewStoryPush(friendUins)
	log.Debugf("end GeneNewStory")
	return
}

func GeneNewStoryPush(uins []int64) (err error) {

	if len(uins) == 0 {
		return
	}

	//往channel里面放有新动态的用户
	for _, uin := range uins {
		im.ChanStoryPush <- uin
	}

	return
}
