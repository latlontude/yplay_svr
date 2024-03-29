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
	"time"
)

//新增回答 发推送
func SendNewAddAnswerPush(uin int64, qid int, answer st.AnswersInfo) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	question, err := getV2Question(qid)
	if err != nil {
		return
	}
	qidOwner := question.OwnerInfo.Uin

	if qidOwner == uin {
		return
	}

	type NewAddAnswerMsg struct {
		Question  st.V2QuestionInfo `json:"question"`
		NewAnswer st.AnswersInfo    `json:"newAnswer"`
		Ts        int64             `json:"ts"`
	}

	var newAddAnswerMsg NewAddAnswerMsg
	newAddAnswerMsg.Question = question
	newAddAnswerMsg.NewAnswer = answer
	newAddAnswerMsg.Ts = time.Now().Unix()
	data, err := json.Marshal(&newAddAnswerMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	//给提问者发送push，告诉ta，ta的提问收到新回答  dataType:17
	descStr := "收到新消息"

	//自己提问自己回答  不需要给自己发通知
	if qidOwner != uin {
		go im.SendV2CommonMsg(serviceAccountUin, qidOwner, 17, dataStr, descStr)
	}

	//给回答过这道题目的人发送push， 告诉ta， ta回答过的题目有了新回答 dataTYpe:14
	uids, err := getAnswerUidByQid(qid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//TODO : uids 应该去重uin
	for _, uid := range uids {
		//不需要给自己发推送   自问自答不重复发
		if uid == uin || uid == qidOwner {
			continue
		}
		//go im.SendV2CommonMsg(serviceAccountUin, uid, 14, dataStr, descStr)
	}

	//同问该问题的人 发推送
	sameAskUidArr, err := GetSameAskUidArr(qid)
	for _, uid := range sameAskUidArr {
		if uid == "" {
			continue
		}
		sameAskUin, err := strconv.ParseInt(uid, 10, 64)

		if sameAskUin == uin {
			continue
		}

		if err != nil {
			log.Errorf(err.Error())
			continue
		}

		log.Errorf("sameAsk ")
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

	sql := fmt.Sprintf(`select ownerUid from  v2answers where qid = %d group by ownerUid`, qid)
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
