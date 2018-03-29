package story

import (
	"api/im"
	"common/constant"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/st"
)

type GeneStoryShareMsgWithTextReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"`
	Uid     int64  `schema:"uid"`  //要分享给的用户
	Text    string `schema:"text"` //聊天消息
}

type GeneStoryShareMsgWithTextRsp struct {
}

type StoryShareMsgWithText struct {
	StoryMsg     st.RetStoryInfo    `json:"storyMsg"`
	Status       int                `json:"status"`       // 1 互为好友，2 storyOwnUin单向添加toUin为好友，3 toUin单向添加storyOwnUin为好友，0 storyOwnUin 和 toUin不是好友，互相之间也没有发送过添加好友请求。
	SendInfo     st.UserProfileInfo `json:"senderInfo"`   //发出分享的人的个人信息
	ReceiverInfo st.UserProfileInfo `json:"receiverInfo"` //接受分享的人的个人信息
	Text         string             `json:"text"`         //聊天消息
}

func doGeneStoryShareMsgWithText(req *GeneStoryShareMsgWithTextReq, r *http.Request) (rsp *GeneStoryShareMsgWithTextRsp, err error) {

	log.Debugf("uin %d, doGeneStoryShareMsgWithTextRsp %+v", req.Uin, req)

	err = GeneStoryShareMsgWithText(req.Uin, req.Uid, req.StoryId, req.Text)
	if err != nil {
		log.Errorf("uin %d, GeneStoryShareMsgWithTextRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GeneStoryShareMsgWithTextRsp{}

	log.Debugf("uin %d, GeneStoryShareMsgWithTextRsp succ, %+v", req.Uin, rsp)

	return
}

func GeneStoryShareMsgWithText(uin, uid, storyId int64, text string) (err error) {
	log.Debugf("start GeneStoryShareMsgWithText uin:%d uid:%d storyId:%d text:%s", uin, uid, storyId, text)

	var shareMsg StoryShareMsgWithText
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_MSG)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", storyId)
	storyVal, err := app.Get(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		err = rest.NewAPIError(constant.E_REDIS_KEY_NO_EXIST, "story expire")
		log.Errorf(err.Error())
		return
	}
	log.Debugf("storyVal:%s", storyVal)

	var story st.StoryInfo
	err = json.Unmarshal([]byte(storyVal), &story)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("story:%+v", story)

	shareMsg.StoryMsg.StoryId = story.StoryId
	shareMsg.StoryMsg.Type = story.Type
	shareMsg.StoryMsg.Text = story.Text
	shareMsg.StoryMsg.Data = story.Data
	shareMsg.StoryMsg.Uin = story.Uin
	shareMsg.StoryMsg.ThumbnailImgUrl = story.ThumbnailImgUrl
	shareMsg.StoryMsg.ViewCnt = story.ViewCnt
	shareMsg.StoryMsg.Ts = story.Ts

	uids := make([]int64, 0)
	uids = append(uids, uin, uid, story.Uin)
	log.Debugf("uids :%+v", uids)

	res, err := st.BatchGetUserProfileInfo(uids)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, u := range uids {
		if _, ok := res[u]; !ok {
			err = rest.NewAPIError(constant.E_USER_NOT_EXIST, "user not exists!")
			log.Errorf(err.Error())
			return
		}
	}

	status, err := getFriendsStatus(story.Uin, uid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	shareMsg.StoryMsg.NickName = res[story.Uin].NickName
	shareMsg.StoryMsg.HeadImgUrl = res[story.Uin].HeadImgUrl

	shareMsg.SendInfo = *res[uin]
	shareMsg.ReceiverInfo = *res[uid]
	shareMsg.Status = status
	shareMsg.Text = text

	log.Debugf("shareMsg:%+v", shareMsg)

	data, err := json.Marshal(&shareMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := fmt.Sprintf("%s 发来新消息", shareMsg.SendInfo.NickName)
	go im.SendStoryShareMsgWithText(uin, uid, dataStr, descStr)

	log.Debugf("end GeneStoryShareMsgWithText")
	return
}
