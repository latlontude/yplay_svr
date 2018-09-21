package v2push

import (
	"api/common"
	"api/im"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"svr/st"
	"time"
)

func SendV2BeCommentPush(from int64, to int64, qid, answerId int, comment st.CommentInfo) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	//兼容旧版本 如果客户端没传qid 去查一遍
	if qid == 0 {
		newQid, _, err := getQidAnswerUidByAnswerId(answerId)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		qid = newQid
	}

	question, err := getV2Question(qid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	answer, err := common.GetV2Answer(answerId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	type BeCommentMsg struct {
		Question   st.V2QuestionInfo `json:"question"`
		MyAnswer   st.AnswersInfo    `json:"myAnswer"`
		NewComment st.CommentInfo    `json:"newComment"`
		Ts         int64             `json:"ts"`
	}

	var beCommentMsg BeCommentMsg
	beCommentMsg.NewComment = comment
	beCommentMsg.Question = question
	beCommentMsg.MyAnswer = answer
	beCommentMsg.Ts = time.Now().Unix()

	data, err := json.Marshal(&beCommentMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16
	if to != from {
		go im.SendV2CommonMsg(serviceAccountUin, to, 16, dataStr, descStr)
	}
}

//被评论 发推送
func SendBeCommentPush(uin int64, answerId int, comment st.CommentInfo) {

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

	answer, err := common.GetV2Answer(answerId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	type BeCommentMsg struct {
		Question   st.V2QuestionInfo `json:"question"`
		MyAnswer   st.AnswersInfo    `json:"myAnswer"`
		NewComment st.CommentInfo    `json:"newComment"`
		Ts         int64             `json:"ts"`
	}

	var beCommentMsg BeCommentMsg
	beCommentMsg.NewComment = comment
	beCommentMsg.Question = question
	beCommentMsg.MyAnswer = answer
	beCommentMsg.Ts = time.Now().Unix()

	data, err := json.Marshal(&beCommentMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给回答者发送push，告诉ta，ta的回答收到了新评论 dataType:16

	if answerUid != uin {
		go im.SendV2CommonMsg(serviceAccountUin, answerUid, 16, dataStr, descStr)
	}
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
