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
type SendLogSettingMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	User  int64  `schema:"user"`
	Act   string `schema:"act"` //act -> debug/info/error/waring/upload
}

//YPLAY后台的发送消息响应
type SendLogSettingMsgRsp struct {
}

func doSendLogSettingMsg(req *SendLogSettingMsgReq, r *http.Request) (rsp *SendLogSettingMsgRsp, err error) {

	log.Errorf("uin %d, SendLogSettingMsgReq %+v", req.Uin, req)

	err = SendLogSettingMsg(req.Uin, req.User, req.Act)
	if err != nil {
		log.Errorf("uin %d, SendLogSettingMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendLogSettingMsgRsp{}

	log.Errorf("uin %d, SendLogSettingMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMLogSettingMsg(uin int64, user int64, act string) (msg IMC2CMsg, err error) {

	var customData IMCustomData
	customData.DataType = 8
	customData.Data = act

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var msgBody IMMsgBody
	msgBody.MsgType = "TIMCustomElem"
	msgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", user)
	msg.MsgBody = []IMMsgBody{msgBody}
	msg.MsgLifeTime = 0

	var offlinePush OfflinePushInfo

	offlinePush.PushFlag = 1
	msg.OfflinePush = offlinePush

	return
}

func SendLogSettingMsg(uin int64, user int64, act string) (err error) {

	if uin == 0 || user == 0 || len(act) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	//操作必须是下面的
	if act != "debug" && act != "info" && act != "error" && act != "warning" && act != "upload" {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMLogSettingMsg(uin, user, act)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendLogSettingMsgReq uin %d, user %d, act %s, msg %+v", uin, user, act, msg)

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

	log.Errorf("SendLogSettingMsgRsp, uin %d, user %d, act %s, rsp %+v", uin, user, act, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
