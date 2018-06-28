package v2push

import (
	"api/im"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"strconv"
	"svr/st"
)

//新增回答 发推送
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

	//同问该问题的人 发推送
	sameAskUidArr, err := GetSameAskUidArr(qid)
	for _, uid := range sameAskUidArr {
		if uid == "" {
			continue
		}
		sameAskUin, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		go im.SendV2CommonMsg(serviceAccountUin, sameAskUin, 14, dataStr, descStr)
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

func SendBeLikedAnswerPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question st.V2QuestionInfo  `json:"question"`
		MyAnswer st.AnswersInfo     `json:"myAnswer"`
		NewLiker st.UserProfileInfo `json:"newLiker"`
	}
	var beLikedMsg BeLikedMsg

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
