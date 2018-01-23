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

//YPLAY后台的发送消息请求包
type SendSubmitQustionApprovedMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	User int64 `schema:"user"`
}

//YPLAY后台的发送消息响应
type SendSubmitQustionApprovedMsgRsp struct {
}

func doSendSubmitQustionApprovedMsg(req *SendSubmitQustionApprovedMsgReq, r *http.Request) (rsp *SendSubmitQustionApprovedMsgRsp, err error) {

	log.Debugf("uin %d, SendSubmitQustionApprovedMsgReq %+v", req.Uin, req)

	err = SendSubmitQustionApprovedMsg(req.User)
	if err != nil {
		log.Errorf("uin %d, SendSubmitQustionApprovedMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendSubmitQustionApprovedMsgRsp{}

	log.Debugf("uin %d, SendSubmitQustionApprovedMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMSubmitQustionApprovedMsg(uin int64) (msg IMC2CMsg, err error) {

	log.Debugf("begin MakeIMSubmitQustionApprovedMsg uin %d", uin)

	var customData IMCustomData
	customData.DataType = 9
	customData.Data = ""

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var newMsgBody IMMsgBody
	newMsgBody.MsgType = "TIMCustomElem"
	newMsgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{newMsgBody}
	msg.MsgLifeTime = 0

	var offlinePush OfflinePushInfo

	offlinePush.PushFlag = 1
	msg.OfflinePush = offlinePush

	return
}

func SendSubmitQustionApprovedMsg(uin int64) (err error) {

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

	msg, err := MakeIMSubmitQustionApprovedMsg(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//log.Errorf("SendNewFeedMsgReq  %+v", msg)

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
		log.Error(err.Error())
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

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
