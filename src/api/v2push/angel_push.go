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
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		BoardId  int                 `json:"boardId"`
		MsgId    int64               `json:"msgId"`
		Ts       int64               `json:"ts"`
	}

	var inviteAngelPush InviteAngelPush
	inviteAngelPush.Operator = inviteInfo
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

//接受邀请 给主天使发邀请
func SendAcceptAngelPush(uin int64, toUin int64) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	AcceptInfo, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}
	type AcceptAngelPush struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		Ts       int64               `json:"ts"`
	}
	var acceptAngelPush AcceptAngelPush
	acceptAngelPush.Operator = AcceptInfo
	acceptAngelPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&acceptAngelPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)
	descStr := "收到新消息"

	log.Debug("descStr:%s  dataStr:%s", descStr, dataStr)
	go im.SendV2CommonMsg(serviceAccountUin, toUin, 27, dataStr, descStr)

	return
}

//主天使 卸任小天使 发送push
func SendDeleteAngelByBigAngelPush(uin int64, toUin int64) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	AcceptInfo, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}
	type DeleteAngelPush struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		Ts       int64               `json:"ts"`
	}

	var deleteAngelPush DeleteAngelPush
	deleteAngelPush.Operator = AcceptInfo
	deleteAngelPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&deleteAngelPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)
	descStr := "收到新消息"

	log.Debug("descStr:%s  dataStr:%s", descStr, dataStr)
	go im.SendV2CommonMsg(serviceAccountUin, toUin, 28, dataStr, descStr)
	return
}

//自己卸任 给主天使发通知
func SendDeleteAngelBySelfPush(uin int64, toUin int64) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	AcceptInfo, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}
	type DeleteAngelPush struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		Ts       int64               `json:"ts"`
	}
	var deleteAngelPush DeleteAngelPush
	deleteAngelPush.Operator = AcceptInfo
	deleteAngelPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&deleteAngelPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)
	descStr := "收到新消息"
	log.Debug("descStr:%s  dataStr:%s", descStr, dataStr)
	go im.SendV2CommonMsg(serviceAccountUin, toUin, 29, dataStr, descStr)
	return
}

//主天使转让
func SendDemiseAngelPush(uin int64, newAngelUin int64, toUin int64) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号

	operator, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}

	newAngelInfo, err := st.GetUserProfileInfo(newAngelUin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}

	type DemiseAngelPush struct {
		Operator     *st.UserProfileInfo `json:"operator"` //操作者
		NewAngelInfo *st.UserProfileInfo `json:"newAngelInfo"`
		Ts           int64               `json:"ts"`
	}

	var demiseAngelPush DemiseAngelPush
	demiseAngelPush.Operator = operator
	demiseAngelPush.NewAngelInfo = newAngelInfo
	demiseAngelPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&demiseAngelPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	dataStr := string(data)
	descStr := "收到新消息"

	log.Debug("descStr:%s  dataStr:%s,uin:%d,touin:%d", descStr, dataStr, uin, toUin)

	dataType := 31 //其他天使
	if newAngelUin == toUin {
		dataType = 30 //新主天使
	}
	go im.SendV2CommonMsg(serviceAccountUin, toUin, dataType, dataStr, descStr)
	return
}

//申请天使
func SendApplyAngelPush(uin int64, boardId int, msgId int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号
	operator, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}

	type ApplyAngelPush struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		BoardId  int                 `json:"boardId"`
		MsgId    int                 `json:"msgId"`
		Ts       int64               `json:"ts"`
	}

	var applyAngelPush ApplyAngelPush
	applyAngelPush.Operator = operator
	applyAngelPush.BoardId = boardId
	applyAngelPush.MsgId = msgId
	applyAngelPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&applyAngelPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	dataStr := string(data)
	descStr := "收到新消息"
	log.Debugf("申请天使 ===> descStr:%s  dataStr:%s, uin:%d ", descStr, dataStr, uin)
	go im.SendV2CommonMsg(102688, serviceAccountUin, 32, dataStr, descStr)

	text1 := "睡着了zZZ,明天早上10点马上回复你!"
	text2 := "本噗10分钟内联系你注册时的电话号码"
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", timeStr)
	timeZero := t.Unix() - 8*3600 //服务器零点转成是8点
	time10 := timeZero + 10*3600
	time22 := timeZero + 22*3600

	now := time.Now().Unix()
	sessionId, err := im.GetSnapSession(serviceAccountUin, uin)
	log.Debugf("time0:%d,time10:%d,time22:%d,now:%d", timeZero, time10, time22, now)
	if now > time10 && now < time22 {
		go im.SendTextMsg(sessionId, text2, serviceAccountUin, uin)
	} else {
		go im.SendTextMsg(sessionId, text1, serviceAccountUin, uin)
	}

	return
}

//审核 天使
func SendCheckApplyPush(uin int64, toUin int64, boardId int, result int) {
	var serviceAccountUin int64
	serviceAccountUin = 100001 //客服号
	if uin != serviceAccountUin {
		log.Debugf("check uin error , is not 1000001 ,uin:%d", uin)
		return
	}
	operator, err := st.GetUserProfileInfo(serviceAccountUin)

	if err != nil {
		log.Errorf("get %d userprofile error :%v ", uin, err)
		return
	}

	type CheckApplyPush struct {
		Operator *st.UserProfileInfo `json:"operator"` //操作者
		Type     int                 `json:"type"`     //1 小天使 通过 2未通过 3主天使
		Ts       int64               `json:"ts"`
	}

	var checkApplyPush CheckApplyPush
	checkApplyPush.Operator = operator
	checkApplyPush.Type = result
	checkApplyPush.Ts = time.Now().Unix()

	data, err := json.Marshal(&checkApplyPush)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	dataStr := string(data)
	descStr := "收到新消息"
	log.Debugf("审核天使 ==== > descStr:%s  dataStr:%s,uin:%d,touin:%d ", descStr, dataStr, uin, toUin)

	go im.SendV2CommonMsg(serviceAccountUin, toUin, 33, dataStr, descStr)

	return
}
