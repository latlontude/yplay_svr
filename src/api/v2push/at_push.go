package v2push

import (
	"api/im"
	_ "common/constant"
	_ "common/rest"
	"encoding/json"
	"svr/st"
	"time"
)

type AtPush struct {
	Avatar   int    `json:"avatar"`
	Uin      int64  `json:"uin"`
	UserName string `json:"username"`
}

type AtPushList struct {
	AtPushList []*AtPush `json:"atPushList"`
}

//被评论 发推送
func SendAtPush(uin int64, pushType int, qid int, info interface{}, ext string) (err error) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	var atPush []*AtPush

	log.Debugf("ext:%s", ext)
	json.Unmarshal([]byte(ext), &atPush)

	for _, info := range atPush {
		log.Debugf("info:%+v", info)
	}

	//if err != nil {
	//	err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
	//	log.Errorf(err.Error())
	//	return
	//}

	//@人数为0  直接返回
	if len(atPush) <= 0 {
		log.Debugf("atPush list is empty  , ext :%s", ext)
		return
	}

	question, err := getV2Question(qid)
	if err != nil {
		return
	}

	type PushMsg struct {
		Question st.V2QuestionInfo `json:"question"`
		Answer   interface{}       `json:"answer"`
		Comment  interface{}       `json:"comment"`
		Reply    interface{}       `json:"reply"`
		Ts       int64             `json:"ts"`
		Type     int               `json:"type"`
	}

	var pushMsg PushMsg
	//@问题
	pushMsg.Question = question
	pushMsg.Ts = time.Now().Unix()
	pushMsg.Type = pushType

	if pushType == 1 {

	} else if pushType == 2 {
		pushMsg.Answer = info
	} else if pushType == 3 {
		pushMsg.Comment = info
	} else if pushType == 4 {
		pushMsg.Reply = info
	}

	for _, atPushInfo := range atPush {
		if atPushInfo.Uin == uin {
			//continue
		}

		data, err := json.Marshal(&pushMsg)
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		dataStr := string(data)
		log.Debugf("dataStr:%s", dataStr)
		descStr := "收到新消息"
		go im.SendV2CommonMsg(serviceAccountUin, atPushInfo.Uin, 23, dataStr, descStr)
	}
	return

}
