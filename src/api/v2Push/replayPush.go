package v2push

import (
	"api/im"
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"svr/st"
)

//回复被点赞 发推送
func SendBeLikedReplyPush(uid int64, qid, likeId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	type BeLikedMsg struct {
		Question st.V2QuestionInfo  `json:"question"`
		MyReply    st.ReplyInfo       `json:"myReply"`
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
