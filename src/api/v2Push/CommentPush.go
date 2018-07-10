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

//被评论 发推送
func SendBeCommentPush(answerId int, comment st.CommentInfo) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	qid, answerUid, err := getQidAnswerUidByAnswerId(answerId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	question, err := getV2Question(qid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	answer, err := getV2Answer(answerId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	type BeCommentMsg struct {
		Question   st.V2QuestionInfo `json:"question"`
		MyAnswer   st.AnswersInfo    `json:"myAnswer"`
		NewComment st.CommentInfo    `json:"newComment"`
	}

	var beCommentMsg BeCommentMsg
	beCommentMsg.NewComment = comment
	beCommentMsg.Question = question
	beCommentMsg.MyAnswer = answer

	data, err := json.Marshal(&beCommentMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16
	go im.SendV2CommonMsg(serviceAccountUin, answerUid, 16, dataStr, descStr)
}

func getQidAnswerUidByAnswerId(answerId int) (qid int, answerUid int64, err error) {

	if answerId == 0 {
		log.Errorf("answerId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qid, ownerUid from  v2answers where answerId = %d`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		rows.Scan(&qid, &answerUid)
	}
	return
}

//评论被点赞 发推送
func SendBeLikedCommentPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question  st.V2QuestionInfo  `json:"question"`
		MyComment st.CommentInfo     `json:"myComment"`
		NewLiker  st.UserProfileInfo `json:"newLiker"`
		Ts       int64               `json:"ts"`
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
	comment, commentOwnerUid, err := getV2Comment(likeId)
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
	}else {
		log.Debugf("commentOwnerUid:u%  uid=%d", commentOwnerUid,uid)
	}

	return
}

//根据commentId 获取commentInfo
//ownerUid 该评论归属者

func getV2Comment(commentId int) (comment st.CommentInfo, ownerUid int64, err error) {
	if commentId == 0 {
		log.Errorf("commentId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select commentId ,answerId,commentContent,ownerUid,commentTs from v2comments where commentId = %d`,
		commentId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var uid int64

	for rows.Next() {
		rows.Scan(&comment.CommentId, &comment.AnswerId, &comment.CommentContent, &uid, &comment.CommentTs)
	}

	ownerUid = uid
	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		comment.OwnerInfo = ui
	}

	return
}
