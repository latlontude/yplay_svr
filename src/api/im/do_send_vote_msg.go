package im

import (
	"bytes"
	"common/constant"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"svr/cache"
	"svr/st"
	"time"
	//"encoding/base64"
)

//IM后台的发送GROUP消息请求包
type IMMsg struct {
	GroupId     string          `json:"GroupId"`
	Random      int             `json:"Random"`
	FromAccount string          `json:"From_Account"`
	MsgBody     []IMMsgBody     `json:"MsgBody"`
	OfflinePush OfflinePushInfo `json:"OfflinePushInfo"`
}

type IMMsgBody struct {
	MsgType    string      `json:"MsgType"`
	MsgContent interface{} `json:"MsgContent"`
}

type IMCustomContent struct {
	Data  string `json:"Data"`
	Desc  string `json:"Desc"`
	Ext   string `json:"Ext"`
	Sound string `json:"Sound"`
}

//两种类型，1是投票 2是投票的第一次回复消息
type IMCustomData struct {
	DataType int    `json:"DataType"`
	Data     string `json:"Data"`
}

type VoteData struct {
	Question     st.QuestionInfo    `json:"QuestionInfo"`
	Options      []st.OptionInfo2   `json:"options"`
	SelIndex     int                `json:"selIndex"`
	SenderInfo   st.UserProfileInfo `json:"senderInfo"`
	ReceiverInfo st.UserProfileInfo `json:"receiverInfo"`
}

type IMSendMsgRsp struct {
	ActionStatus string `json:"ActionStatus"`
	ErrorInfo    string `json:"ErrorInfo"`
	ErrorCode    int    `json:"ErrorCode"`
	MsgTime      int    `json:"MsgTime"`
	MsgSeq       int    `json:"MsgSeq"`
}

type OfflinePushInfo struct {
	PushFlag int         `json:"PushFlag"`
	Desc     string      `json:"Desc"`
	Ext      string      `json:"Ext"`
	Apns     ApnsInfo    `json:"ApnsInfo"`
	Ands     AndroidInfo `json:"AndroidInfo"`
}

type ApnsInfo struct {
	BadgeMode int    `json:"BadgeMode"`
	Sound     string `json:"Sound"`
	Title     string `json:"Title"`
	SubTitle  string `json:"SubTitle"`
}

type AndroidInfo struct {
	Title string `json:"Title"`
}

type NotifyExtInfo struct {
	NotifyType int    `json:"notifyType"`
	Content    string `json:"content"`
}

type NotifyExtIMContent struct {
	SessionId string `json:"sessionId"`
	Status    int    `json:"status"`
}

//YPLAY后台的发送消息请求包
type SendVoteMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	QId          int    `schema:"qid"`
	VoteToUin    int64  `schema:"voteToUin"`
	OptionStr    string `schema:"optionStr"`
	VoteRecordId int64  `schema:"voteRecordId"`
}

//YPLAY后台的发送消息响应
type SendVoteMsgRsp struct {
}

func doSendVoteMsg(req *SendVoteMsgReq, r *http.Request) (rsp *SendVoteMsgRsp, err error) {

	log.Debugf("uin %d, SendVoteMsgReq %+v", req.Uin, req)

	err = SendVoteMsg(req.Uin, req.QId, req.VoteToUin, req.OptionStr, req.VoteRecordId)
	if err != nil {
		log.Errorf("uin %d, SendVoteMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendVoteMsgRsp{}

	log.Debugf("uin %d, SendVoteMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMVoteMsg(uin int64, qid int, voteToUin int64, optionStr string, groupId string) (msg IMMsg, err error) {

	var options []st.OptionInfo2
	err = json.Unmarshal([]byte(optionStr), &options)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, err.Error())
		log.Errorf(err.Error())
		return
	}

	if len(options) != constant.ENUM_OPTION_BATCH_SIZE {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid options size")
		log.Errorf(err.Error())
		return
	}

	voteToNickName := ""
	selIndex := 0
	for i, v := range options {
		if v.Uin == voteToUin {
			selIndex = i + 1
			voteToNickName = v.NickName
			break
		}
	}

	res, err := st.BatchGetUserProfileInfo([]int64{uin, voteToUin})
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(res) != 2 {
		err = rest.NewAPIError(constant.E_USER_NOT_EXIST, "user not exists!")
		log.Errorf(err.Error())
		return
	}

	var voteData VoteData
	voteData.Question = *cache.QUESTIONS[qid]
	voteData.Options = options
	voteData.SelIndex = selIndex
	voteData.SenderInfo = *res[uin]
	voteData.ReceiverInfo = *res[voteToUin]

	var customData IMCustomData

	customData.DataType = 1
	cd, _ := json.Marshal(voteData)
	customData.Data = string(cd)

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var voteMsgBody IMMsgBody
	voteMsgBody.MsgType = "TIMCustomElem"
	voteMsgBody.MsgContent = customContent

	msg.GroupId = groupId
	msg.Random = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{voteMsgBody}

	var offlinePush OfflinePushInfo

	offlinePush.Desc = "你收到了一个匿名投票"

	ui, err1 := st.GetUserProfileInfo(uin)
	if err1 == nil {
		genderStr := "男生"
		if ui.Gender == 2 {
			genderStr = "女生"
		}

		//@{收到push人的用户名}，神秘{投票人年级}{投票人性别}对你说了真心话( ⁼̴̀ .̫ ⁼̴́ )✧
		offlinePush.Desc = fmt.Sprintf("@%s，神秘%s%s对你说了真心话( ⁼̴̀ .̫ ⁼̴́ )✧", voteToNickName, st.GetGradeDescBySchool(ui.SchoolType, ui.Grade), genderStr)

	} else {
		log.Errorf(err1.Error())
	}

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
	offlinePush.Apns = ApnsInfo{0, "", "噗噗", ""}
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendVoteMsg(uin int64, qid int, voteToUin int64, optionStr string, voteRecordId int64) (err error) {

	if uin == 0 || voteToUin == 0 || qid == 0 || len(optionStr) == 0 || voteRecordId == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	if voteToUin == 0 {
		return
	}

	if uin == voteToUin {
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	groupName := cache.QUESTIONS[qid].QText

	groupId, err := CreateGroup(uin, voteToUin, voteRecordId, groupName)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMVoteMsg(uin, qid, voteToUin, optionStr, groupId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendVoteMsgReq uin %d, qid %d, voteToUin %d, voteRecordId %d, req %+v", uin, qid, voteToUin, voteRecordId, msg)

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

	log.Errorf("IMSendVoteMsgRsp uin %d, qid %d, voteToUin %d, voteRecordId %d, rsp %+v", uin, qid, voteToUin, voteRecordId, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
