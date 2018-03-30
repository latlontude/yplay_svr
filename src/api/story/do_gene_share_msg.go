package story

import (
	"api/im"
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/st"
)

type GeneStoryShareMsgReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"`
	Uid     int64  `schema:"uid"` //要分享给的用户
}

type GeneStoryShareMsgRsp struct {
}

type StoryShareMsg struct {
	StoryMsg     st.RetStoryInfo    `json:"storyMsg"`
	Status       int                `json:"status"`       // 1 互为好友，2 storyOwnUin单向添加toUin为好友，3 toUin单向添加storyOwnUin为好友，0 storyOwnUin 和 toUin不是好友，互相之间也没有发送过添加好友请求。
	SendInfo     st.UserProfileInfo `json:"senderInfo"`   //发出分享的人的个人信息
	ReceiverInfo st.UserProfileInfo `json:"receiverInfo"` //接受分享的人的个人信息
}

func doGeneStoryShareMsg(req *GeneStoryShareMsgReq, r *http.Request) (rsp *GeneStoryShareMsgRsp, err error) {

	log.Debugf("uin %d, doGeneStoryShareMsgRsp %+v", req.Uin, req)

	err = GeneStoryShareMsg(req.Uin, req.Uid, req.StoryId)
	if err != nil {
		log.Errorf("uin %d, GeneStoryShareMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GeneStoryShareMsgRsp{}

	log.Debugf("uin %d, GeneStoryShareMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func GeneStoryShareMsg(uin, uid, storyId int64) (err error) {
	log.Debugf("start GeneStoryShareMsg uin:%d uid:%d storyId:%d", uin, uid, storyId)

	var shareMsg StoryShareMsg
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

	log.Debugf("shareMsg:%+v", shareMsg)

	data, err := json.Marshal(&shareMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := fmt.Sprintf("%s 发来新消息", shareMsg.SendInfo.NickName)
	go im.SendStoryShareMsg(uin, uid, dataStr, descStr)

	log.Debugf("end GeneStoryShareMsg")
	return
}

func getFriendsStatus(uid1, uid2 int64) (status int, err error) {
	log.Debugf("start getFriendsStatus uid1:%d uid2:%d", uid1, uid2)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select fromUin, toUin, status from addFriendMsg where (fromUin = %d and toUin = %d) or (fromUin = %d and toUin = %d)`, uid1, uid2, uid2, uid1)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	var fromUin int64
	var toUin int64
	var stat int
	for rows.Next() {
		rows.Scan(&fromUin, &toUin, &stat)
		if stat == 1 {
			status = constant.ENUM_SNS_STATUS_IS_FRIEND //互为好友
		} else if fromUin == uid1 {
			status = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND // uid1 单向添加uid2为好友，但uid2未同意
		} else {
			status = constant.ENUM_SNS_STATUS_FRIEND_HAS_INVAITE_ME //uid2 单向添加uid1为好友，但uid1未同意
		}
		break
	}

	if fromUin == 0 && toUin == 0 && stat == 0 {
		status = constant.ENUM_SNS_STATUS_NOT_FRIEND //非好友
	}

	if uid1 == uid2 {
		status = constant.ENUM_SNS_STATUS_IS_FRIEND //互为好友 把自己的消息分享给自己
	}

	log.Debugf("end getFriendsStatus status:%d", status)
	return
}
