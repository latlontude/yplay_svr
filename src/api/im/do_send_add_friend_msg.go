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

type IMC2CMsg struct {
	SyncOtherMachine int             `json:"SyncOtherMachine"`
	FromAccount      string          `json:"From_Account"`
	ToAccount        string          `json:"To_Account"`
	MsgLifeTime      int             `json:"MsgLifeTime"`
	MsgRandom        int             `json:"MsgRandom"`
	MsgTimeStamp     int             `json:"MsgTimeStamp"`
	MsgBody          []IMMsgBody     `json:"MsgBody"`
	OfflinePush      OfflinePushInfo `json:"OfflinePushInfo"`
}

//YPLAY后台的发送消息请求包
type SendAddFriendMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Uin1 int64 `schema:"uin1"`
	Uin2 int64 `schema:"uin2"`
}

//YPLAY后台的发送消息响应
type SendAddFriendMsgRsp struct {
}

func doSendAddFriendMsg(req *SendAddFriendMsgReq, r *http.Request) (rsp *SendAddFriendMsgRsp, err error) {

	log.Debugf("uin %d, SendAddFriendMsgReq %+v", req.Uin, req)

	err = SendAddFriendMsg(req.Uin1, req.Uin2)
	if err != nil {
		log.Errorf("uin %d, SendAddFriendMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendAddFriendMsgRsp{}

	log.Debugf("uin %d, SendAddFriendMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMAddFriendMsg(uin1 int64, uin2 int64) (msg IMC2CMsg, err error) {

	ui, err := st.GetUserProfileInfo(uin1)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var customData IMCustomData
	customData.DataType = 3
	customData.Data = fmt.Sprintf("%s", ui.NickName)

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var addFriendMsgBody IMMsgBody
	addFriendMsgBody.MsgType = "TIMCustomElem"
	addFriendMsgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", uin2)
	msg.MsgBody = []IMMsgBody{addFriendMsgBody}
	msg.MsgLifeTime = 0 //若消息只发在线用户，不想保存离线，则该字段填0。

	var offlinePush OfflinePushInfo

	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_ADD_FRIEND
	extInfo.Content = fmt.Sprintf("%s", ui.NickName)

	se, _ := json.Marshal(extInfo)

	content := fmt.Sprintf("%s:同学～加个好友呗(*/ω＼*)", ui.NickName)

	offlinePush.PushFlag = 0
	offlinePush.Desc = content
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{0, "", "", ""}
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendAddFriendMsg(uin1 int64, uin2 int64) (err error) {

	if uin1 == 0 || uin2 == 0 || uin1 == uin2 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMAddFriendMsg(uin1, uin2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("SendAddFriendMsgReq uin1 %d, uin2 %d, req %+v", uin1, uin2, msg)

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

	log.Debugf("SendAddFriendMsgRsp uin1 %d, uin2 %d, rsp %+v", uin1, uin2, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
