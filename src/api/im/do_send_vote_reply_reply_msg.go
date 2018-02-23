package im

import (
	"bytes"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"svr/st"
	"time"
	//"encoding/base64"
)

type VoteReplyReplyData struct {
	Reply        string             `json:"reply"`
	ReplyReply   string             `json:"replyreply"`
	Question     st.QuestionInfo    `json:"QuestionInfo"`
	Options      []st.OptionInfo2   `json:"options"`
	SelIndex     int                `json:"selIndex"`
	SenderInfo   st.UserProfileInfo `json:"senderInfo"`
	ReceiverInfo st.UserProfileInfo `json:"receiverInfo"`
}

//YPLAY后台的发送消息请求包
type SendVoteReplyReplyMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SessionId  string `schema:"sessionId"`
	Reply      string `schema:"reply"`
	ReplyReply string `schema:"replyReply"`
}

//YPLAY后台的发送消息响应
type SendVoteReplyReplyMsgRsp struct {
}

func doSendVoteReplyReplyMsg(req *SendVoteReplyReplyMsgReq, r *http.Request) (rsp *SendVoteReplyReplyMsgRsp, err error) {

	log.Debugf("uin %d, SendVoteReplyReplyMsgReq %+v", req.Uin, req)

	err = SendVoteReplyReplyMsg(req.Uin, req.SessionId, req.Reply, req.ReplyReply)
	if err != nil {
		log.Errorf("uin %d, SendVoteReplyReplyMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendVoteReplyReplyMsgRsp{}

	log.Debugf("uin %d, SendVoteReplyReplyMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMVoteReplyReplyMsg(uin int64, sessionId string, newSessionId string, reply string, replyReply string) (msg IMMsg, err error) {

	//从老的会话找到原有的投票消息
	record, err := st.GetVoteRecordInfoByImSessionId(sessionId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if uin != record.Uin {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "permision deny")
		log.Errorf(err.Error())
		return
	}

	selIndex := 0
	for i, v := range record.Options {
		if v.Uin == record.VoteToUin {
			selIndex = i + 1
			break
		}
	}

	var voteData VoteReplyReplyData
	voteData.Reply = reply
	voteData.ReplyReply = replyReply
	voteData.Question = st.QuestionInfo{record.QId, record.QText, record.QIconUrl, 0, 0, 0, 0, 0, 0, "", 0, "", 0, "", 0, "", 0}
	for _, option := range record.Options {
		voteData.Options = append(voteData.Options, *option)
	}
	voteData.SelIndex = selIndex
	voteData.ReceiverInfo = st.UserProfileInfo{}
	voteData.ReceiverInfo.Uin = record.VoteToUin
	voteData.ReceiverInfo.NickName = record.VoteToNickName
	voteData.ReceiverInfo.HeadImgUrl = record.VoteToHeadImgUrl
	voteData.ReceiverInfo.Gender = record.VoteToGender
	voteData.ReceiverInfo.Age = record.VoteToAge
	voteData.ReceiverInfo.Grade = record.VoteToGrade
	voteData.ReceiverInfo.SchoolId = record.VoteToSchoolId
	voteData.ReceiverInfo.SchoolType = record.VoteToSchoolType
	voteData.ReceiverInfo.SchoolName = record.VoteToSchoolName

	voteData.SenderInfo = st.UserProfileInfo{}
	voteData.SenderInfo.Uin = record.Uin
	voteData.SenderInfo.NickName = record.NickName
	voteData.SenderInfo.HeadImgUrl = record.HeadImgUrl
	voteData.SenderInfo.Gender = record.Gender
	voteData.SenderInfo.Age = record.Age
	voteData.SenderInfo.Grade = record.Grade
	voteData.SenderInfo.SchoolId = record.SchoolId
	voteData.SenderInfo.SchoolType = record.SchoolType
	voteData.SenderInfo.SchoolName = record.SchoolName

	var customData IMCustomData

	customData.DataType = 1000
	cd, _ := json.Marshal(voteData)
	customData.Data = string(cd)

	var customContent IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var voteReplyMsgBody IMMsgBody
	voteReplyMsgBody.MsgType = "TIMCustomElem"
	voteReplyMsgBody.MsgContent = customContent

	msg.GroupId = newSessionId
	msg.Random = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{voteReplyMsgBody}

	var offlinePush OfflinePushInfo

	senderNickName := voteData.SenderInfo.NickName

	var extContent NotifyExtIMContent
	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_IM

	extContent.SessionId = newSessionId
	extContent.Status = 2
	sb, _ := json.Marshal(extContent)
	extInfo.Content = string(sb)

	se, _ := json.Marshal(extInfo)

	offlinePush.PushFlag = 0
	offlinePush.Desc = fmt.Sprintf("%s 发来新消息", senderNickName)
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{0, "", "噗噗", ""}
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendVoteReplyReplyMsg(uin int64, sessionId string, reply string, replyReply string) (err error) {

	if uin == 0 || len(sessionId) == 0 || len(replyReply) == 0 {
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
	isFriend, user, err := CheckImSessionUserIsFriend2(uin, sessionId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//如果不是朋友，就不发消息了
	if isFriend == 0 {
		log.Errorf("IMSendVoteReplyReplyMsgReq uin %d, sessionId %s, session users are not friends", uin, sessionId)
		err = rest.NewAPIError(constant.E_IM_NOT_FRIEND, "not friends")
		log.Error(err.Error())
		return
		//return
	}

	//先看固定会话是否存在，如果不存在，则创建
	newSessionId, err := CreateSnapChatSesson(uin, user)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendVoteReplyReplyMsg CreateSnapChatSesson, uin %d, user %d, orgSessionId %s, SnapChatSession %s", uin, user, sessionId, newSessionId)

	msg, err := MakeIMVoteReplyReplyMsg(uin, sessionId, newSessionId, reply, replyReply)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendVoteReplyReplyMsgReq uin %d, sessionId %s, req %+v", uin, newSessionId, msg)

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

	log.Errorf("IMSendVoteReplyReplyMsgRsp uin %d, sessionId %s, rsp %+v", uin, newSessionId, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}

func CheckImSessionUserIsFriend2(uin int64, sessionId string) (isFriend int, user int64, err error) {

	isFriend = 0

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}
	sql := fmt.Sprintf(`select uin, voteToUin from voteRecords where imSessionId = "%s"`, sessionId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	var uin1 int64
	var uin2 int64

	find := false
	for rows.Next() {

		rows.Scan(&uin1, &uin2)
		find = true
		break
	}

	if !find {
		return
	}

	isFriend, err = st.IsFriend(uin1, uin2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if uin1 == uin {
		user = uin2
	} else {
		user = uin1
	}

	return
}
