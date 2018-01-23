package account

import (
	"common/constant"
	//"common/env"
	"common/env"
	"common/myredis"
	"common/rest"
	"common/sms"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type SendSmsReq struct {
	Phone string `schema:"phone"`
}

type SendSmsRsp struct {
	Code string `json:"code"`
}

func doSendSms(req *SendSmsReq, r *http.Request) (rsp *SendSmsRsp, err error) {

	log.Debugf("phone %s, SendSmsReq %+v", req.Phone, req)

	code, err := SendLoginSms(req.Phone)
	if err != nil {
		log.Errorf("phone %s, SendSmsRsp error, %s", req.Phone, err.Error())
		return
	}

	rsp = &SendSmsRsp{code}

	log.Debugf("phone %s, SendSmsRsp succ, %+v", req.Phone, rsp)

	return
}

func SendLoginSms(phone string) (code string, err error) {

	if !sms.IsValidPhone(phone) {
		err = rest.NewAPIError(constant.E_INVALID_PHONE, "phone number invalid")
		log.Error(err.Error())
		return
	}

	randor := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := randor.Intn(9000) + 1000
	randNumStr := fmt.Sprintf("%d", randNum)

	//为了DEBUG方便，返回给客户端，上线之后要去掉
	code = fmt.Sprintf("%d", randNum)

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SMS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%s", phone)
	err = app.SetEx(keyStr, randNumStr, uint32(env.Config.Sms.TTL)) //63秒
	if err != nil {
		log.Error(err.Error())
		return
	}

	if phone == "18682235582" || phone == "13480970139" {
		return
	}

	err = sms.SendPhoneCode(phone, randNumStr)
	if err != nil {
		err = rest.NewAPIError(constant.E_SMS_SEND_ERR, "send sms error,"+err.Error())
		log.Error(err.Error())
		return
	}

	return
}
