package v2push

import (
	"api/im"
	"encoding/json"
	"svr/st"
	"time"
)

//经验弹加入 删除 通知

//新增回答 发推送
func SendAddAnswerIdInExpPush(uin int64, qid int, labelId int, answerId int) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	question, err := getV2Question(qid)
	if err != nil {
		return
	}
	answer, err := getV2Answer(answerId)
	if err != nil {
		return
	}

	toUid := answer.OwnerInfo.Uin

	operatorInfo, err := st.GetUserProfileInfo(uin)

	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}

	type ExpPushMsg struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		Question st.V2QuestionInfo   `json:"question"`
		Answer   st.AnswersInfo      `json:"answer"`
		LabelId  int                 `json:"labelId"`
		Ts       int64               `json:"ts"`
	}

	var expPushMsg ExpPushMsg
	expPushMsg.Question = question
	expPushMsg.Answer = answer
	expPushMsg.LabelId = labelId
	expPushMsg.Ts = time.Now().Unix()
	expPushMsg.Operator = operatorInfo

	data, err := json.Marshal(&expPushMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	//给提问者发送push，告诉ta，ta的提问收到新回答  dataType:17
	descStr := "收到新消息"

	//自己提问自己回答  不需要给自己发通知

	log.Debug("descStr:%s  dataStr:%s", descStr, dataStr)
	go im.SendV2CommonMsg(serviceAccountUin, toUid, 24, dataStr, descStr)

	return
}

//删除经验弹 发推送
func SendDelAnswerIdInExpPush(uin int64, qid int, answerId int, labelId int) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	question, err := getV2Question(qid)
	if err != nil {
		return
	}
	answer, err := getV2Answer(answerId)
	if err != nil {
		return
	}

	toUid := answer.OwnerInfo.Uin

	operatorInfo, err := st.GetUserProfileInfo(uin)

	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}
	type ExpPushMsg struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		Question st.V2QuestionInfo   `json:"question"`
		Answer   st.AnswersInfo      `json:"answer"`
		LabelId  int                 `json:"labelId"`
		Ts       int64               `json:"ts"`
	}

	var expPushMsg ExpPushMsg
	expPushMsg.Question = question
	expPushMsg.Answer = answer
	expPushMsg.LabelId = labelId
	expPushMsg.Ts = time.Now().Unix()
	expPushMsg.Operator = operatorInfo

	data, err := json.Marshal(&expPushMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	//给提问者发送push，告诉ta，ta的提问收到新回答  dataType:17
	descStr := "收到新消息"

	//自己提问自己回答  不需要给自己发通知

	log.Debug("descStr:%s  dataStr:%s", descStr, dataStr)
	go im.SendV2CommonMsg(serviceAccountUin, toUid, 25, dataStr, descStr)

	return
}
