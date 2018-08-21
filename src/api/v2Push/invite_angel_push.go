package v2push

import (
	"api/im"
	"encoding/json"
	"svr/st"
	"time"
)

//邀请
func SendInviteAngelPush(uin int64, toUin int64, boardId int, msgId int64) {

	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	inviteInfo, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}

	type InviteAngelPush struct {
		InviteInfo *st.UserProfileInfo `json:"inviteInfo"` //操作者
		BoardId    int                 `json:"boardId"`
		MsgId      int64               `json:"msgId"`
		Ts         int64               `json:"ts"`
	}

	var inviteAngelPush InviteAngelPush
	inviteAngelPush.InviteInfo = inviteInfo
	inviteAngelPush.BoardId = boardId
	inviteAngelPush.MsgId = msgId
	inviteAngelPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&inviteAngelPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)

	log.Debugf("dataStr:%s", dataStr)

	//给提问者发送push，告诉ta，ta的提问收到新回答  dataType:17
	descStr := "收到新消息"

	//自己提问自己回答  不需要给自己发通知

	log.Debug("descStr:%s  dataStr:%s", descStr, dataStr)
	go im.SendV2CommonMsg(serviceAccountUin, toUin, 26, dataStr, descStr)

	return
}
