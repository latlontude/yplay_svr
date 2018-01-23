package account

import (
	"common/constant"
	"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"common/sms"
	"common/token"
	"fmt"
	"net/http"
	//"strings"
	"svr/st"
)

type Login2Req struct {
	Phone  string `schema:"phone"`
	Code   string `schema:"code"`
	Uuid   int64  `schema:"uuid"`
	Device string `schema:"device"`
	Os     string `schema:"os"`
	AppVer string `schema:"appVer"`
}

type Login2Rsp struct {
	HasCheckInviteCode int    `json:"hasCheckInviteCode"` //是否校验过邀请码
	Uin                int64  `json:"uin"`
	Token              string `json:"token"`
	Ver                int    `json:"ver"`
	IsNewUser          int    `json:"isNewUser"`

	Info *st.UserProfileInfo `json:"info"`
}

func doLogin2(req *Login2Req, r *http.Request) (rsp *Login2Rsp, err error) {

	//isIphone := false

	log.Debugf("phone %s, Login2Req %+v", req.Phone, req)

	/*
		if _, ok := r.Header["X-Wns-Deviceinfo"]; ok {

			wns_device_info := strings.Join(r.Header["X-Wns-Deviceinfo"], ",")

			if strings.Contains(wns_device_info, "iOS") || strings.Contains(wns_device_info, "iPhone") {
				isIphone = true
			}
		}
	*/

	rsp, err = Login2(req.Phone, req.Code, req.Uuid, req.Device, req.Os, req.AppVer)
	if err != nil {
		log.Errorf("phone %s, Login2Rsp error, %s", req.Phone, err.Error())
		return
	}

	log.Debugf("phone %s, Login2Rsp succ, %+v", req.Phone, rsp)

	//if isIphone {
	//	rsp.HasCheckInviteCode = 1
	//}

	return
}

func Login2(phone string, code string, uuid int64, device, os, appVer string) (rsp *Login2Rsp, err error) {

	//2017-11-01 17:02
	if len(phone) == 0 || len(code) == 0 || uuid < constant.ENUM_DEVICE_UUID_MIN {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	if !sms.IsValidPhone(phone) {
		err = rest.NewAPIError(constant.E_INVALID_PHONE, "phone number invalid")
		log.Error(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SMS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%s", phone)
	valStr, err := app.Get(keyStr) //code 缓存63秒

	if (phone != "18682235582" && phone != "13480970139") && err != nil {
		err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "code not exist in redis")
		log.Error(err.Error())
		return
	}

	//短信验证码不能进行再次使用
	app.Del(keyStr)

	//该手机号码不校验验证码
	if phone == "18682235582" || phone == "13480970139" {
		if len(code) != 4 {
			err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "invalid code")
			log.Error(err.Error())
			return
		}
	} else if code != valStr {
		err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "invalid code")
		log.Error(err.Error())
		return
	}

	uin, isNewUser, err := GetUinByPhone(phone)
	if err != nil {
		log.Error(err.Error())
		return
	}

	ttl := env.Config.Token.TTL
	token, err := token.GeneToken(uin, env.Config.Token.VER, ttl, uuid, device, os, appVer)
	if err != nil {
		log.Error(err.Error())
		return
	}

	app, err = myredis.GetApp(constant.ENUM_REDIS_APP_TOKEN)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = app.SetEx(fmt.Sprintf("%d", uin), token, uint32(ttl))
	if err != nil {
		log.Error(err.Error())
		return
	}

	info, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//是否校验过邀请码
	hasCheckInviteCode, err := PhoneInviteCodeHasCheck(phone)

	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &Login2Rsp{hasCheckInviteCode, uin, token, env.Config.Token.VER, isNewUser, info}

	return
}

func PhoneInviteCodeHasCheck(phone string) (pass int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	q := fmt.Sprintf(`select phone from invitecode where phone = "%s"`, phone)

	rows, err := inst.Query(q)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	pass = 0
	for rows.Next() {
		var tmp string
		rows.Scan(&tmp)
		pass = 1
	}

	return

}
