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

type SendNewFeedMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	User int64 `schema:"user"`
}

//YPLAY后台的发送消息响应
type SendNewFeedMsgRsp struct {
}

func doSendNewFeedMsg(req *SendNewFeedMsgReq, r *http.Request) (rsp *SendNewFeedMsgRsp, err error) {

	log.Debugf("uin %d, SendNewFeedMsgReq %+v", req.Uin, req)

	err = SendNewFeedMsg(req.User)
	if err != nil {
		log.Errorf("uin %d, SendNewFeedMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendNewFeedMsgRsp{}

	log.Debugf("uin %d, SendNewFeedMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMNewFeedMsg(uin int64) (msg IMC2CMsg, err error) {

	log.Debugf("begin MakeIMNewFeedMsg uin %d", uin)

	var customData IMCustomData
	customData.DataType = 5
	customData.Data = ""

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
	msg.ToAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{newFeedMsgBody}
	msg.MsgLifeTime = 0

	var offlinePush OfflinePushInfo

	offlinePush.PushFlag = 1
	msg.OfflinePush = offlinePush

	return
}

func SendNewFeedMsg(uin int64) (err error) {

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

	msg, err := MakeIMNewFeedMsg(uin)
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

	//log.Errorf("SendNewFeedMsgRsp %+v", rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
