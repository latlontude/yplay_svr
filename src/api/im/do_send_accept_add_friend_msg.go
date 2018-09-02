package im

import (
	"bytes"
	"common/constant"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"svr/st"
	"time"
	//"encoding/base64"
)

//YPLAY后台的发送消息请求包
type SendAcceptAddFriendMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Uin1 int64 `schema:"uin1"`
	Uin2 int64 `schema:"uin2"`
}

//YPLAY后台的发送消息响应
type SendAcceptAddFriendMsgRsp struct {
}

func doSendAcceptAddFriendMsg(req *SendAcceptAddFriendMsgReq, r *http.Request) (rsp *SendAcceptAddFriendMsgRsp, err error) {

	log.Debugf("uin %d, SendAcceptAddFriendMsgReq %+v", req.Uin, req)

	err = SendAcceptAddFriendMsg(req.Uin1, req.Uin2)
	if err != nil {
		log.Errorf("uin %d, SendAcceptAddFriendMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendAcceptAddFriendMsgRsp{}

	log.Debugf("uin %d, SendAcceptAddFriendMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMAcceptAddFriendMsg(uin1 int64, uin2 int64) (msg IMC2CMsg, err error) {

	ui, err := st.GetUserProfileInfo(uin2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	ui.Ts = int(time.Now().Unix()) //成为好友的时间

	byteData, _ := json.Marshal(ui)

	var customData IMCustomData
	customData.DataType = 7
	customData.Data = string(byteData)

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var newFeedMsgBody IMMsgBody
	newFeedMsgBody.MsgType = "TIMCustomElem"
	newFeedMsgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", uin1)
	msg.MsgBody = []IMMsgBody{newFeedMsgBody}

	//若消息只发在线用户，不想保存离线，则该字段填0。这里填写非0 表示需要产生离线消息 目标是作为加好友之后能在对方的最近联系人列表出现
	msg.MsgLifeTime = 604800

	var offlinePush OfflinePushInfo

	offlinePush.PushFlag = 1
	msg.OfflinePush = offlinePush

	return
}

func SendAcceptAddFriendMsg(uin1 int64, uin2 int64) (err error) {

	if uin1 == 0 || uin2 == 0 || uin1 == uin2 && uin1 != 100001 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMAcceptAddFriendMsg(uin1, uin2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendAcceptFriendMsgReq uin1 %d, uin2 %d, msg %+v", uin1, uin2, msg)

	data, err := json.Marshal(&msg)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/openim/sendmsg?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
		sig, constant.ENUM_IM_IDENTIFIER_ADMIN, constant.ENUM_IM_SDK_APPID, time.Now().Unix())

	hrsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	body, err := ioutil.ReadAll(hrsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	var rsp IMSendMsgRsp

	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendAcceptFriendMsgRsp, uin1 %d, uin2 %d, rsp %+v", uin1, uin2, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}

func SendBeFriendsStartChatMsg(fromUin, toUin int64) (err error) {
	log.Debugf("start SendBeFriendsStartChatMsg fromUin:%d, toUin:%d", fromUin, toUin)

	sessionId, err := GetSnapSession(fromUin, toUin)
	if err != nil {
		log.Errorf(err.Error())
		log.Errorf("faied to get sessionId")
		return
	}

	text := "我们已成为好友啦，开始聊天吧ᕕ( ᐛ )ᕗ"
	SendTextMsg(sessionId, text, fromUin, toUin)
	log.Debugf("end SendBeFriendsStartChatMsg")
	return
}
