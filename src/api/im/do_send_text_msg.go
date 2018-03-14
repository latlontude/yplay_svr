package im

import (
	"bytes"
	"common/constant"
	//"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"svr/st"
	"time"
	//"encoding/base64"
)

type IMTextMsgContent struct {
	Text string `json:"Text"`
}

//YPLAY后台的发送文本消息请求包
type SendTextMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SessionId   string `schema:"sessionId"`
	Text        string `schema:"text"`
	FromAccount int64  `schema:"fromAccount"`
	ToAccount   int64  `schema:"toAccount"`
}

//YPLAY后台的发送消息响应
type SendTextMsgRsp struct {
}

func doSendTextMsg(req *SendTextMsgReq, r *http.Request) (rsp *SendTextMsgRsp, err error) {

	log.Debugf("uin %d, SendVoteReplyMsgReq %+v", req.Uin, req)

	err = SendTextMsg(req.SessionId, req.Text, req.FromAccount, req.ToAccount)
	if err != nil {
		log.Errorf("uin %d, SendTextMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendTextMsgRsp{}

	log.Debugf("uin %d, SendTextMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMTextMsg(sessionId string, text string, fromAccount, toAccount int64) (msg IMMsg, err error) {

	ui, err := st.GetUserProfileInfo(fromAccount)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var body IMMsgBody
	body.MsgType = "TIMTextElem"
	body.MsgContent = IMTextMsgContent{text}

	msg.GroupId = sessionId
	msg.Random = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", fromAccount)
	msg.MsgBody = []IMMsgBody{body}

	var offlinePush OfflinePushInfo

	var extContent NotifyExtIMContent
	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_IM

	extContent.SessionId = sessionId
	extContent.Status = 2 //! 实名之后的会话消息
	sb, _ := json.Marshal(extContent)
	extInfo.Content = string(sb)

	se, _ := json.Marshal(extInfo)

	offlinePush.PushFlag = 0
	offlinePush.Desc = fmt.Sprintf("%s 发来新消息", ui.NickName)
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{0, "", "", ""}
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendTextMsg(sessionId string, text string, fromAccount, toAccount int64) (err error) {

	if len(sessionId) == 0 || len(text) == 0 || fromAccount == 0 || toAccount == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//判断会话的两个人是否还是朋友
	isFriend, err := st.IsFriend(fromAccount, toAccount)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//如果不是朋友，就不发消息了
	if isFriend == 0 {
		log.Errorf("IMSendTextMsgReq fromUin %d, toUin %d, sessionId %s, session users are not friends", fromAccount, toAccount, sessionId)
		err = rest.NewAPIError(constant.E_IM_NOT_FRIEND, "not friends")
		log.Error(err.Error())
		return
	}

	msg, err := MakeIMTextMsg(sessionId, text, fromAccount, toAccount)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendTextMsgReq fromUin %d, toUin %d, sessionId %s, req %+v", fromAccount, toAccount, sessionId, msg)

	data, err := json.Marshal(&msg)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/group_open_http_svc/send_group_msg?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
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

	log.Errorf("IMSendTextMsgRsp fromUin %d, toUin %d, sessionId %s, rsp %+v", fromAccount, toAccount, sessionId, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
