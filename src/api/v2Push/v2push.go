package v2push

import (
	"api/im"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"svr/st"
)

var log = env.NewLogger("v2push")



func SendNewAddAnswerPush(qid int, answer st.AnswersInfo) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	question, err := getV2Question(qid)
	if err != nil {
		return
	}
	qidOwner := question.OwnerInfo.Uin

	type NewAddAnswerMsg struct {
		Question  st.V2QuestionInfo `json:"question"`
		NewAnswer st.AnswersInfo    `json:"newAnswer"`
	}

	var newAddAnswerMsg NewAddAnswerMsg
	newAddAnswerMsg.Question = question
	newAddAnswerMsg.NewAnswer = answer

	data, err := json.Marshal(&newAddAnswerMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	//给提问者发送push，告诉ta，ta的提问收到新回答  dataType:17
	descStr := "收到新消息"
	go im.SendV2CommonMsg(serviceAccountUin, qidOwner, 17, dataStr, descStr)

	//给回答过这道题目的人发送push， 告诉ta， ta回答过的题目有了新回答 dataTYpe:14
	uids, err := getAnswerUidByQid(qid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uid := range uids {
		go im.SendV2CommonMsg(serviceAccountUin, uid, 14, dataStr, descStr)
	}

	return
}

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

func SendBeLikePush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question st.V2QuestionInfo  `json:"question"`
		MyAnswer st.AnswersInfo     `json:"myAnswer"`
		NewLiker st.UserProfileInfo `json:"newLiker"`
	}
	var beLikedMsg BeLikedMsg

	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		beLikedMsg.NewLiker = *ui
	}

	question, err := getV2Question(qid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	beLikedMsg.Question = question

	answer, err := getV2Answer(likeId)
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
	go im.SendV2CommonMsg(serviceAccountUin, answerUid, 15, dataStr, descStr)

	return
}


/**
删除提问发推送
 */
func SendBeDeletePush(operatorUid int64 , uid int64 , reason string, deleteType int) {

	type BeDeleteMsg struct {
		Type 		int  					`json:"type"`		// type: 1:提问被删除 2：回答被删除 3：评论被删除
		Operator    st.UserProfileInfo 		`json:"operator"`
		Ts  		int64					`json:"ts"`
		Reason 		string					`json:"reason"`
	}

	var deleteMsg BeDeleteMsg;

	if operatorUid > 0 {
		ui, err1 := st.GetUserProfileInfo(operatorUid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		deleteMsg.Operator = *ui
	}

	deleteMsg.Type 		 = deleteType
	deleteMsg.Ts         = 0
	deleteMsg.Reason 	 = reason

	data, err := json.Marshal(&deleteMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给question 所属着发推送
	go im.SendV2CommonMsg(operatorUid, uid, 18, dataStr, descStr)

	return
}


func getV2Question(qid int) (question st.V2QuestionInfo, err error) {
	if qid == 0 {
		log.Errorf("qid is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qid, qTitle, qContent, qImgUrls, ownerUid, isAnonymous, createTs, modTs  from  v2questions where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var ownerUid int64

	for rows.Next() {
		rows.Scan(&question.Qid, &question.QTitle, &question.QContent, &question.QImgUrls, &ownerUid, &question.IsAnonymous, &question.CreateTs, &question.ModTs)
	}

	if ownerUid > 0 {
		ui, err1 := st.GetUserProfileInfo(ownerUid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		question.OwnerInfo = ui
	}
	return
}

func getV2Answer(answerId int) (answer st.AnswersInfo, err error) {
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

	sql := fmt.Sprintf(`select qid, answerId, ownerUid, answerContent, answerImgUrls, answerTs from  v2answers where answerId = %d`, answerId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var uid int64
	for rows.Next() {
		rows.Scan(&answer.Qid, &answer.AnswerId, &uid, &answer.AnswerContent, &answer.AnswerImgUrls, &answer.AnswerTs)
	}

	if uid > 0 {
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		answer.OwnerInfo = ui
	}

	return
}

func getAnswerUidByQid(qid int) (uids []int64, err error) {

	if qid == 0 {
		log.Errorf("qid is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select ownerUid from  v2answers where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		uids = append(uids, uid)
	}
	return
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
