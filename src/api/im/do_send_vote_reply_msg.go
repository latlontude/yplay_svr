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

type VoteReplyData struct {
	Content      string             `json:"content"`
	Question     st.QuestionInfo    `json:"QuestionInfo"`
	Options      []st.OptionInfo2   `json:"options"`
	SelIndex     int                `json:"selIndex"`
	SenderInfo   st.UserProfileInfo `json:"senderInfo"`
	ReceiverInfo st.UserProfileInfo `json:"receiverInfo"`
}

//YPLAY后台的发送消息请求包
type SendVoteReplyMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SessionId string `schema:"sessionId"`
	Content   string `schema:"content"`
}

//YPLAY后台的发送消息响应
type SendVoteReplyMsgRsp struct {
}

func doSendVoteReplyMsg(req *SendVoteReplyMsgReq, r *http.Request) (rsp *SendVoteReplyMsgRsp, err error) {

	log.Debugf("uin %d, SendVoteReplyMsgReq %+v", req.Uin, req)

	err = SendVoteReplyMsg(req.Uin, req.SessionId, req.Content)
	if err != nil {
		log.Errorf("uin %d, SendVoteReplyMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SendVoteReplyMsgRsp{}

	log.Debugf("uin %d, SendVoteReplyMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMVoteReplyMsg(uin int64, sessionId string, content string) (msg IMMsg, err error) {

	record, err := st.GetVoteRecordInfoByImSessionId(sessionId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if uin != record.VoteToUin {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "permision deny")
		log.Errorf(err.Error())
		return
	}

	selIndex := 0
	for i, v := range record.Options {
		if v.Uin == uin {
			selIndex = i + 1
			break
		}
	}

	var voteData VoteReplyData
	voteData.Content = content
	voteData.Question = st.QuestionInfo{record.QId, record.QText, record.QIconUrl, 0, 0, 0, 0, 0, 0, "", 0, "", 0, "", 0, "", 0}
	for _, option := range record.Options {
		voteData.Options = append(voteData.Options, *option)
	}
	voteData.SelIndex = selIndex
	voteData.SenderInfo = st.UserProfileInfo{}
	voteData.SenderInfo.Uin = record.VoteToUin
	voteData.SenderInfo.NickName = record.VoteToNickName
	voteData.SenderInfo.HeadImgUrl = record.VoteToHeadImgUrl
	voteData.SenderInfo.Gender = record.VoteToGender
	voteData.SenderInfo.Age = record.VoteToAge
	voteData.SenderInfo.Grade = record.VoteToGrade
	voteData.SenderInfo.SchoolId = record.VoteToSchoolId
	voteData.SenderInfo.SchoolType = record.VoteToSchoolType
	voteData.SenderInfo.SchoolName = record.VoteToSchoolName

	voteData.ReceiverInfo = st.UserProfileInfo{}
	voteData.ReceiverInfo.Uin = record.Uin
	voteData.ReceiverInfo.NickName = record.NickName
	voteData.ReceiverInfo.HeadImgUrl = record.HeadImgUrl
	voteData.ReceiverInfo.Gender = record.Gender
	voteData.ReceiverInfo.Age = record.Age
	voteData.ReceiverInfo.Grade = record.Grade
	voteData.ReceiverInfo.SchoolId = record.SchoolId
	voteData.ReceiverInfo.SchoolType = record.SchoolType
	voteData.ReceiverInfo.SchoolName = record.SchoolName

	var customData IMCustomData

	customData.DataType = 2
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

	msg.GroupId = sessionId
	msg.Random = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []IMMsgBody{voteReplyMsgBody}

	var offlinePush OfflinePushInfo

	senderNickName := voteData.SenderInfo.NickName

	var extContent NotifyExtIMContent
	var extInfo NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_IM

	extContent.SessionId = sessionId
	extContent.Status = 1
	sb, _ := json.Marshal(extContent)
	extInfo.Content = string(sb)

	se, _ := json.Marshal(extInfo)

	offlinePush.PushFlag = 0
	offlinePush.Desc = fmt.Sprintf("%s 发来新消息", senderNickName)
	offlinePush.Ext = string(se)
	offlinePush.Apns = ApnsInfo{0, "", "", ""}
	offlinePush.Ands = AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendVoteReplyMsg(uin int64, sessionId string, content string) (err error) {

	if uin == 0 || len(sessionId) == 0 || len(content) == 0 {
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
	isFriend, err := CheckImSessionUserIsFriend(uin, sessionId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//如果不是朋友，就不发消息了
	if isFriend == 0 {
		log.Errorf("IMSendVoteReplyMsgReq uin %d, sessionId %s, session users are not friends", uin, sessionId)
		err = rest.NewAPIError(constant.E_IM_NOT_FRIEND, "not friends")
		log.Error(err.Error())
		return
		//return
	}

	msg, err := MakeIMVoteReplyMsg(uin, sessionId, content)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("IMSendVoteReplyMsgReq uin %d, sessionId %s, req %+v", uin, sessionId, msg)

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

	log.Errorf("IMSendVoteReplyMsgRsp uin %d, sessionId %s, rsp %+v", uin, sessionId, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}

func CheckImSessionUserIsFriend(uin int64, sessionId string) (isFriend int, err error) {

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

	return
}
