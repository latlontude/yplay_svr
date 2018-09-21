package v2push

import (
	"api/common"
	"api/im"
	"encoding/json"
	"svr/st"
	"time"
)

//评论被点赞 发推送
func SendBeLikedCommentPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question  st.V2QuestionInfo  `json:"question"`
		MyComment st.CommentInfo     `json:"myComment"`
		NewLiker  st.UserProfileInfo `json:"newLiker"`
		Ts        int64              `json:"ts"`
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

	//被点赞的评论  根据likeId(commentId) 查找commentInfo
	comment, commentOwnerUid, err := common.GetV2Comment(likeId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	beLikedMsg.MyComment = comment

	data, err := json.Marshal(&beLikedMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给评论者发送push，告诉ta，ta的评论被点赞 dataType:19
	//自己赞自己 就不要发通知了
	if commentOwnerUid != uid {
		go im.SendV2CommonMsg(serviceAccountUin, commentOwnerUid, 19, dataStr, descStr)
	} else {
		log.Debugf("commentOwnerUid:u%  uid=%d", commentOwnerUid, uid)
	}

	return
}

func SendBeLikedAnswerPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question st.V2QuestionInfo  `json:"question"`
		MyAnswer st.AnswersInfo     `json:"myAnswer"`
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

	//被点赞的问题
	answer, err := common.GetV2Answer(likeId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	answerUid := answer.OwnerInfo.Uin
	beLikedMsg.MyAnswer = answer

	data, err := json.Marshal(&beLikedMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回答者发送push，告诉ta，ta的回答被点赞 dataType:15
	//自己赞自己 就不要发通知了
	if answerUid != uid {
		go im.SendV2CommonMsg(serviceAccountUin, answerUid, 15, dataStr, descStr)
	}

	return
}

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
	reply, replyOwnerUid, err := common.GetV2Reply(likeId)
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

//评论被点赞 发推送
func SendBeLikedQuestionPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question st.V2QuestionInfo  `json:"question"`
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

	data, err := json.Marshal(&beLikedMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给评论者发送push，告诉ta，ta的评论被点赞 dataType:19
	//自己赞自己 就不要发通知了
	if question.OwnerInfo.Uin != uid && question.OwnerInfo.Uin != 0 {
		go im.SendV2CommonMsg(serviceAccountUin, question.OwnerInfo.Uin, 19, dataStr, descStr)
	} else {
		log.Debugf("OwnerUid:u%  uid=%d", question.OwnerInfo.Uin, uid)
	}

	return
}
