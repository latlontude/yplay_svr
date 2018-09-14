package v2push

import (
	"api/im"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"svr/st"
	"time"
)

//回复被点赞 发推送
func SendBeLikedReplyPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question st.V2QuestionInfo  `json:"question"`
		MyReply  st.ReplyInfo       `json:"myReply"`
		NewLiker st.UserProfileInfo `json:"newLiker"`
		Ts       int64              `json:"ts"`
	}

	var beLikedMsg BeLikedMsg
	beLikedMsg.Ts = time.Now().Unix()

	//点赞人info
	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		beLikedMsg.NewLiker = *ui
	}

	//被点赞所属问题
	question, err := getV2Question(qid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	beLikedMsg.Question = question

	//根据replyId(likeId)  查找replyInfo
	reply, replyOwnerUid, err := getV2Reply(likeId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	beLikedMsg.MyReply = reply

	data, err := json.Marshal(&beLikedMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回复者发送push，告诉ta，ta的回复被点赞 dataType:20
	if replyOwnerUid != uid {
		go im.SendV2CommonMsg(serviceAccountUin, replyOwnerUid, 20, dataStr, descStr)
	}

	return
}

//回复被回复
func SendReplyBeReplyPush(uin int64, qid, answerId, commentId, replyId int, newReply st.ReplyInfo) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号
	question, _ := getV2Question(qid)
	answer, _ := getV2Answer(answerId)
	comment, _, _ := getV2Comment(commentId)
	reply, replyOwnerUid, err := getV2Reply(replyId)
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
	answer, _ := getV2Answer(answerId)
	comment, commentOwnerUid, err := getV2Comment(commentId)
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

//replyId replyInfo
//ownerUid 该回复归属者

func getV2Reply(replyId int) (reply st.ReplyInfo, ownerUid int64, err error) {
	if replyId == 0 {
		log.Errorf("replyId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select replyId,replyContent,fromUid,toUid,replyTs  from v2replys where replyId = %d`, replyId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	var fromUid int64
	var toUid int64

	for rows.Next() {
		rows.Scan(&reply.ReplyId, &reply.ReplyContent, &fromUid, &toUid, &reply.ReplyTs)
	}

	//被点赞对象
	ownerUid = toUid

	if fromUid > 0 {
		ui, err1 := st.GetUserProfileInfo(fromUid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		reply.ReplyFromUserInfo = ui
	}

	if toUid > 0 {
		ui, err1 := st.GetUserProfileInfo(toUid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		reply.ReplyToUserInfo = ui
	}

	return
}
