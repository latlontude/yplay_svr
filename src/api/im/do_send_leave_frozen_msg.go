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
	//"svr/st"
	//"encoding/base64"
)

//YPLAY后台的发送消息请求包
type SendLeaveFrozenMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	User int64 `schema:"user"`
}

//YPLAY后台的发送消息响应
type SendLeaveFrozenMsgRsp struct {
}

func doSendLeaveFrozenMsg(req *SendLeaveFrozenMsgReq, r *http.Request) (rsp *SendLeaveFrozenMsgRsp, err error) {

	log.Debugf("uin %d, SendLeaveFrozenMsgReq %+v", req.Uin, req)

	content := "开始新一轮投票吧(๑‾ ꇴ ‾๑)"

	err = SendLeaveFrozenMsg(req.User, content)
	if err != nil {
		log.Errorf("uin %d, SendLeaveFrozenMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendLeaveFrozenMsgRsp{}

	log.Debugf("uin %d, SendLeaveFrozenMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMLeaveFrozenMsg(uin int64, content string) (msg IMC2CMsg, err error) {

	// ui, err := st.GetUserProfileInfo(uin)
	// if err != nil{
	//     log.Error(err.Error())
	//     return
	// }

	log.Debugf("begin MakeIMLeaveFrozenMsg uin %d", uin)

	var customData IMCustomData
	customData.DataType = 4
	customData.Data = ""

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var leaveFrozenMsgBody IMMsgBody
	leaveFrozenMsgBody.MsgType = "TIMCustomElem"
	leaveFrozenMsgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{leaveFrozenMsgBody}
	msg.MsgLifeTime = 604800

	var offlinePush OfflinePushInfo

	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_LEAVE_FROZEN
	extInfo.Content = ""

	se, _ := json.Marshal(extInfo)

	//content := fmt.Sprintf("开始新一轮投票吧(๑‾ ꇴ ‾๑)")

	offlinePush.PushFlag = 0
	offlinePush.Desc = content
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{1, "", "", ""} //badge不计数
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendLeaveFrozenMsg(uin int64, content string) (err error) {

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

	msg, err := MakeIMLeaveFrozenMsg(uin, content)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendLeaveFrozenMsgReq uin %d, msg %+v", uin, msg)

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

	log.Errorf("SendLeaveFrozenMsgRsp uin %d, rsp %+v", uin, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
