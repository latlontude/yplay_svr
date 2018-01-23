package sns

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/rest"
	"common/sms"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/st"
	//"common/myredis"
	"time"
)

type InviteFriendsBySmsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Friends string `schema:"friends"` //好友ID的手机号列表,给我的好友的非注册APP用户发短信,客户端base64编码避免明文
}

type InviteFriendsBySmsRsp struct {
}

func doInviteFriendsBySms(req *InviteFriendsBySmsReq, r *http.Request) (rsp *InviteFriendsBySmsRsp, err error) {

	log.Debugf("uin %d, InviteFriendsBySmsReq %+v", req.Uin, req)

	//先不发送短信
	err = InviteFriendsBySms(req.Uin, req.Friends)
	if err != nil {
		log.Errorf("uin %d, InviteFriendsBySmsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &InviteFriendsBySmsRsp{}

	log.Debugf("uin %d, InviteFriendsBySmsRsp succ, %+v", req.Uin, rsp)

	return
}

func InviteFriendsBySms(uin int64, data string) (err error) {

	if uin == 0 || len(data) == 0 {
		return
	}

	decodeData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err)
		return
	}

	var phones []string
	err = json.Unmarshal([]byte(decodeData), &phones)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err)
		return
	}

	if len(phones) == 0 {
		return
	}

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		return
	}

	//如果邀请其他用户，则判断当前是否冷冻状态，如果是，则解冻
	go st.LeaveFrozenStatusByInviteFriend(uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	stmt, err := inst.Prepare(`insert ignore into inviteFriendSms(id, uin, phone, status, ts) values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	msg := fmt.Sprintf("%s的%s，希望加你为好友, 一起来玩呗! http://x.y.z", ui.SchoolName, ui.NickName)

	if len(ui.SchoolName) == 0 {
		msg = fmt.Sprintf("%s，希望加你为好友, 一起来玩呗! http://x.y.z", ui.NickName)
	}

	for _, phone := range phones {

		if !sms.IsValidPhone(phone) {
			continue
		}

		//不实际发送短信
		//go sms.SendPhoneMsg(phone, msg)
		fmt.Sprintf("phone %s, msg %+v", phone, msg)

		_, err = stmt.Exec(0, uin, phone, 0, ts)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}
	}

	return
}
