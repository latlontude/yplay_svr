package v2push

import (
	"api/common"
	"api/im"
	"encoding/json"
	"svr/st"
	"time"
)

//回复被回复
func SendReplyBeReplyPush(uin int64, qid, answerId, commentId, replyId int, newReply st.ReplyInfo) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号
	question, _ := getV2Question(qid)
	answer, _ := common.GetV2Answer(answerId)
	comment, _, _ := common.GetV2Comment(commentId)
	reply, replyOwnerUid, err := common.GetV2Reply(replyId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	type BeReplyMsg struct {
		Question st.V2QuestionInfo `json:"question"`
		Answer   st.AnswersInfo    `json:"answer"`
		Comment  st.CommentInfo    `json:"comment"`
		Reply    st.ReplyInfo      `json:"reply"`
		NewReply st.ReplyInfo      `json:"newReply"`
		Ts       int64             `json:"ts"`
	}

	var beReplyMsg BeReplyMsg
	beReplyMsg.Question = question
	beReplyMsg.Answer = answer
	beReplyMsg.Comment = comment
	beReplyMsg.Reply = reply
	beReplyMsg.NewReply = newReply
	beReplyMsg.Ts = time.Now().Unix()

	data, err := json.Marshal(&beReplyMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回答者发送push，告诉ta，ta的回复收到了新回复,dataType:22
	log.Errorf("sendUin:%d,replyOwnerUid:%d", serviceAccountUin, replyOwnerUid)

	if uin != replyOwnerUid {
		go im.SendV2CommonMsg(serviceAccountUin, replyOwnerUid, 22, dataStr, descStr)
	}
}

//评论被回复
func SendCommentBeReplyPush(uin int64, qid, answerId, commentId int, reply st.ReplyInfo) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	question, _ := getV2Question(qid)
	answer, _ := common.GetV2Answer(answerId)
	comment, commentOwnerUid, err := common.GetV2Comment(commentId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	type BeReplyMsg struct {
		Question st.V2QuestionInfo `json:"question"`
		Answer   st.AnswersInfo    `json:"answer"`
		Comment  st.CommentInfo    `json:"comment"`
		NewReply st.ReplyInfo      `json:"newReply"`
		Ts       int64             `json:"ts"`
	}

	var beReplyMsg BeReplyMsg
	beReplyMsg.Question = question
	beReplyMsg.Answer = answer
	beReplyMsg.Comment = comment
	beReplyMsg.NewReply = reply
	beReplyMsg.Ts = time.Now().Unix()

	data, err := json.Marshal(&beReplyMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16
	log.Debugf("sendUin:%d,commentOwnerUid:%d", serviceAccountUin, commentOwnerUid)
	if commentOwnerUid != uin {
		go im.SendV2CommonMsg(serviceAccountUin, commentOwnerUid, 21, dataStr, descStr)
	}
}
