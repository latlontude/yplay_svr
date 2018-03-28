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

func MakeStoryShareMsg(fromUin int64, groupId, data, descStr string) (msg IMMsg, err error) {
	log.Debugf("start MakeStoryShareMsg groupId:%s", groupId)

	var customData IMCustomData
	customData.DataType = 11
	customData.Data = data

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var shareMsgBody IMMsgBody
	shareMsgBody.MsgType = "TIMCustomElem"
	shareMsgBody.MsgContent = customContent

	msg.GroupId = groupId
	msg.Random = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", fromUin)
	msg.MsgBody = []IMMsgBody{shareMsgBody}

	var offlinePush OfflinePushInfo

	offlinePush.Desc = descStr

	//构造ext信息，客户端通过notifyType来区分不同场景的push，然后来跳转
	var extContent NotifyExtIMContent
	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_IM

	extContent.SessionId = groupId
	extContent.Status = 0
	sb, _ := json.Marshal(extContent)
	extInfo.Content = string(sb)

	se, _ := json.Marshal(extInfo)

	offlinePush.PushFlag = 0
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{0, "", "", ""}
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush
	log.Debugf("end MakeStoryShareMsg")
	return
}

func SendStoryShareMsg(fromUin, toUin int64, shareData, descStr string) (err error) {

	log.Debugf("start SendStoryShareMsg fromUin:%d, toUin:%d", fromUin, toUin)
	if fromUin == 0 || toUin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	groupId, err := GetSnapSession(fromUin, toUin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeStoryShareMsg(fromUin, groupId, shareData, descStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendStoryShareMsgReq uin %d, req %+v", fromUin, msg)

	data, err := json.Marshal(&msg)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_STORY_SHARE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/group_open_http_svc/send_group_msg?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
		sig, constant.ENUM_IM_IDENTIFIER_ADMIN, constant.ENUM_IM_SDK_APPID, time.Now().Unix())

	hrsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_STORY_SHARE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	body, err := ioutil.ReadAll(hrsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_STORY_SHARE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	var rsp IMSendMsgRsp

	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_STORY_SHARE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendStoryShareMsgRsp uin %d, rsp %+v", fromUin, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_STORY_SHARE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}
	log.Debugf("end SendStoryShareMsg")
	return
}
