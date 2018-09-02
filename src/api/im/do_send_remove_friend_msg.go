package im

import (
	"bytes"
	"common/constant"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"svr/st"
	"time"
	//"encoding/base64"
)

//YPLAY后台的发送消息请求包
type SendRemoveFriendMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Uin1 int64 `schema:"uin1"`
	Uin2 int64 `schema:"uin2"`
}

//YPLAY后台的发送消息响应
type SendRemoveFriendMsgRsp struct {
}

func doSendRemoveFriendMsg(req *SendRemoveFriendMsgReq, r *http.Request) (rsp *SendRemoveFriendMsgRsp, err error) {

	log.Errorf("uin %d, SendRemoveFriendMsgRsp %+v", req.Uin, req)

	err = SendRemoveFriendMsg(req.Uin1, req.Uin2)
	if err != nil {
		log.Errorf("uin %d, SendRemoveFriendMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendRemoveFriendMsgRsp{}

	log.Errorf("uin %d, SendRemoveFriendMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMRemoveFriendMsg(uin1 int64, uin2 int64) (msg IMC2CMsg, err error) {

	log.Debugf("begin MakeIMRemoveFriendMsg uin1 %d, uin2 %d", uin1, uin2)

	var customData IMCustomData
	customData.DataType = 6
	customData.Data = fmt.Sprintf(`{"uin":%d,"ts":%d}`, uin1, time.Now().Unix()) //uin1 主动解除uin2的好友关系，消息发送给uin2

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
	msg.ToAccount = fmt.Sprintf("%d", uin2)
	msg.MsgBody = []IMMsgBody{newFeedMsgBody}

	//若消息只发在线用户，不想保存离线，则该字段填0。这里填写非0 表示需要产生离线消息 目标是作为加好友之后能在对方的最近联系人列表出现
	msg.MsgLifeTime = 604800

	var offlinePush OfflinePushInfo

	offlinePush.PushFlag = 1
	msg.OfflinePush = offlinePush

	return
}

func SendRemoveFriendMsg(uin1 int64, uin2 int64) (err error) {

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

	msg, err := MakeIMRemoveFriendMsg(uin1, uin2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendRemoveFriendMsgReq uin1 %d, uin2 %d, req %+v", uin1, uin2, msg)

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

	log.Errorf("SendRemoveFriendMsgRsp uin1 %d, uin2 %d, rsp %+v", uin1, uin2, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}

func RemoveFriendS2S(uin1 int64, uin2 int64) (err error) {

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMRemoveFriendMsg(uin1, uin2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendRemoveFriendMsgReq uin1 %d, uin2 %d, req %+v", uin1, uin2, msg)

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

	log.Errorf("SendRemoveFriendMsgRsp uin1 %d, uin2 %d, rsp %+v", uin1, uin2, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
