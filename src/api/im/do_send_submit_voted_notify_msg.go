package im

import (
	"bytes"
	"common/constant"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SendSubmitVotedNotifyMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	User  int64  `schema:"user"`
}

type SendSubmitVotedNotifyMsgRsp struct {
}

func doSendSubmitVotedNotifyMsg(req *SendSubmitVotedNotifyMsgReq, r *http.Request) (rsp *SendSubmitVotedNotifyMsgRsp, err error) {

	log.Debugf("uin %d, SendSubmitVotedNotifyMsgReq %+v", req.Uin, req)

	err = SendSubmitVotedNotifyMsg(req.User)
	if err != nil {
		log.Errorf("uin %d, SendSubmitVotedNotifyMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendSubmitVotedNotifyMsgRsp{}

	log.Debugf("uin %d, SendSubmitVotedNotifyMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMSubmitVotedNotifyMsg(uin int64) (msg IMC2CMsg, err error) {

	log.Debugf("begin MakeIMSubmitVotedNotifyMsg uin %d", uin)

	var customData IMCustomData
	customData.DataType = 10
	customData.Data = ""

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var submitVotedNotifyMsgBody IMMsgBody
	submitVotedNotifyMsgBody.MsgType = "TIMCustomElem"
	submitVotedNotifyMsgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{submitVotedNotifyMsgBody}
	msg.MsgLifeTime = 604800

	var offlinePush OfflinePushInfo

	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_SUBMIT_ADD_NEW_HOT
	extInfo.Content = ""

	se, _ := json.Marshal(extInfo)

	content := fmt.Sprintf("同校同年级的人获得你投稿题目的投票")

	offlinePush.PushFlag = 1 // 不离线推送
	offlinePush.Desc = content
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{1, "", "投稿题有新动态", ""} //badge不计数
	offlinePush.Ands = AndroidInfo{""}

	msg.OfflinePush = offlinePush

	log.Debugf("end MakeIMSubmitVotedNotifyMsg uin %d", uin)

	return
}

func SendSubmitVotedNotifyMsg(uin int64) (err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMSubmitVotedNotifyMsg(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendSubmitVotedNotifyMsgReq uin %d, msg %+v", uin, msg)

	data, err := json.Marshal(&msg)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_NEW_HOT_NOTIFY_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/openim/sendmsg?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
		sig, constant.ENUM_IM_IDENTIFIER_ADMIN, constant.ENUM_IM_SDK_APPID, time.Now().Unix())

	hrsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_NEW_HOT_NOTIFY_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	body, err := ioutil.ReadAll(hrsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_NEW_HOT_NOTIFY_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	var rsp IMSendMsgRsp

	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_NEW_HOT_NOTIFY_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendSubmitVotedNotifyMsgRsp uin %d, rsp %+v", uin, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_NEW_HOT_NOTIFY_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
