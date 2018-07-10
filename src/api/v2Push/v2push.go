package v2push

import (
	"api/im"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"strings"
	"svr/st"
	"time"
)

var log = env.NewLogger("v2push")

/**
	删除提问 发推送
*/
func SendBeDeletePush(operatorUid int64, uid int64, reason string, deleteType int,deleteData string) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号
	type BeDeleteMsg struct {
		Type     int                `json:"type"` // type: 1:提问被删除 2：回答被删除 3：评论被删除 4:回复被删除
		Operator st.UserProfileInfo `json:"operator"`
		Ts       int64              `json:"ts"`
		Reason   string             `json:"reason"`
		Data     string             `json:"data"`
	}

	var deleteMsg BeDeleteMsg

	if operatorUid > 0 {
		ui, err1 := st.GetUserProfileInfo(operatorUid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		deleteMsg.Operator = *ui
	}

	deleteMsg.Type = deleteType
	deleteMsg.Ts = time.Now().Unix()
	deleteMsg.Reason = reason
	deleteMsg.Data = deleteData

	data, err := json.Marshal(&deleteMsg)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	descStr := "收到新消息"

	//给question 所属着发推送
	go im.SendV2CommonMsg(serviceAccountUin, uid, 18, dataStr, descStr)

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








/**
获取同问该问题的uid
 */
func GetSameAskUidArr(qid int) (uids []string, err error) {
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

	sql := fmt.Sprintf(`select sameAskUid from  v2questions where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var sameAskUid string
	for rows.Next() {
		rows.Scan(&sameAskUid)
	}
	uids = strings.Split(sameAskUid,",")

	return
}
