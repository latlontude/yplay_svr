package sms

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type PhoneInfo struct {
	NationCode string `json:"nationcode"`
	Mobile     string `json:"mobile"`
}

type SmsReqBody struct {
	Phone  PhoneInfo `json:"tel"`
	TplId  uint32    `json:"tpl_id"`
	Params []string  `json:"params"`
	Sig    string    `json:"sig"`
	Ts     uint32    `json:"time"`
	Extend string    `json:"extend"`
	Ext    string    `json:"ext"`
}

type SmsRspBody struct {
	Result int32  `schema:"result"`
	ErrMsg string `schema:"errmsg"`
	Ext    string `schema:"ext"`
	Sid    string `schema:"sid"`
	Fee    int32  `schema:"fee"`
}

func IsValidPhone(phone string) (ok bool) {

	reg1_str := "^1(3[0-9]|4[57]|5[0-35-9]|7[01678]|8[0-9])\\d{8}$"
	/**
	 * 中国移动：China Mobile
	 * 134,135,136,137,138,139,147,150,151,152,157,158,159,170,178,182,183,184,187,188
	 */

	reg2_str := "^1(3[4-9]|4[7]|5[0-27-9]|7[0]|7[8]|8[2-478])\\d{8}$"
	/**
	 * 中国联通：China Unicom
	 * 130,131,132,145,152,155,156,1709,171,176,185,186
	 */
	reg3_str := "^1(3[0-2]|4[5]|5[56]|709|7[1]|7[6]|8[56])\\d{8}$"
	/**
	 * 中国电信：China Telecom
	 * 133,134,153,1700,177,180,181,189
	 */

	reg4_str := "^1(7[3]|6[6]|9[0-9])\\d{8}$"

	ok, _ = regexp.MatchString(reg1_str, phone)
	if ok {
		return
	}

	ok, _ = regexp.MatchString(reg2_str, phone)
	if ok {
		return
	}

	ok, _ = regexp.MatchString(reg3_str, phone)
	if ok {
		return
	}

	ok, _ = regexp.MatchString(reg4_str, phone)
	if ok {
		return
	}

	return false
}

func SendPhoneMsg(phone string, text1, text2 string, minute string) (err error) {

	const SMS_SDK_APPID = 1400031527
	const SMS_APPKEY = "a0a26597a1d8c60ac486b1b33359345a"
	const SMS_TPL_ID = 20545

	ts := uint32(time.Now().Unix())

	url := fmt.Sprintf("https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=%d&random=%d", SMS_SDK_APPID, ts)

	s := fmt.Sprintf("appkey=%s&random=%d&time=%d&mobile=%s", SMS_APPKEY, ts, ts, phone)

	h := sha256.New()
	h.Write([]byte(s))
	sig := fmt.Sprintf("%x", h.Sum(nil))

	var req SmsReqBody

	req.Phone.NationCode = "86"
	req.Phone.Mobile = phone
	req.TplId = SMS_TPL_ID
	//req.Type = 0
	//req.Msg = fmt.Sprintf

	req.Sig = string(sig)
	req.Ts = ts
	req.Extend = ""
	req.Ext = ""

	req.Params = []string{text1, text2, minute}

	d, err := json.Marshal(req)
	if err != nil {
		return
	}

	rsp, err := http.Post(url, "application/json", strings.NewReader(string(d)))
	if err != nil {
		return
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	var ret SmsRspBody

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return
	}

	if ret.Result != 0 {
		err = errors.New(ret.ErrMsg)
		return
	}

	return
}

func SendPhoneMsgByTemplate(phone string, params []string, tplId uint32) (err error) {

	const SMS_SDK_APPID = 1400031527
	const SMS_APPKEY = "a0a26597a1d8c60ac486b1b33359345a"

	if len(phone) == 0 || len(params) == 0 || tplId == 0 {
		return
	}

	ts := uint32(time.Now().Unix())

	url := fmt.Sprintf("https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=%d&random=%d", SMS_SDK_APPID, ts)

	s := fmt.Sprintf("appkey=%s&random=%d&time=%d&mobile=%s", SMS_APPKEY, ts, ts, phone)

	h := sha256.New()
	h.Write([]byte(s))
	sig := fmt.Sprintf("%x", h.Sum(nil))

	var req SmsReqBody

	req.Phone.NationCode = "86"
	req.Phone.Mobile = phone
	req.TplId = tplId

	req.Sig = string(sig)
	req.Ts = ts
	req.Extend = ""
	req.Ext = ""
	req.Params = params

	d, err := json.Marshal(req)
	if err != nil {
		return
	}

	rsp, err := http.Post(url, "application/json", strings.NewReader(string(d)))
	if err != nil {
		return
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	var ret SmsRspBody

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return
	}

	if ret.Result != 0 {
		err = errors.New(ret.ErrMsg)
		return
	}

	return
}

func SendPhoneCode(phone string, text string) (err error) {

	const SMS_SDK_APPID = 1400031527
	const SMS_APPKEY = "a0a26597a1d8c60ac486b1b33359345a"
	const SMS_TPL_ID = 20545

	ts := uint32(time.Now().Unix())

	url := fmt.Sprintf("https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=%d&random=%d", SMS_SDK_APPID, ts)

	s := fmt.Sprintf("appkey=%s&random=%d&time=%d&mobile=%s", SMS_APPKEY, ts, ts, phone)

	h := sha256.New()
	h.Write([]byte(s))
	sig := fmt.Sprintf("%x", h.Sum(nil))

	var req SmsReqBody

	req.Phone.NationCode = "86"
	req.Phone.Mobile = phone
	req.TplId = SMS_TPL_ID
	//req.Type = 0
	//req.Msg = fmt.Sprintf

	req.Sig = string(sig)
	req.Ts = ts
	req.Extend = ""
	req.Ext = ""

	//req.Params = []string{"[YPLAY]您的验证码", text, "1"}
	req.Params = []string{"【噗噗】登录验证码", text, "3"}

	d, err := json.Marshal(req)
	if err != nil {
		return
	}

	rsp, err := http.Post(url, "application/json", strings.NewReader(string(d)))
	if err != nil {
		return
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	var ret SmsRspBody

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return
	}

	if ret.Result != 0 {
		err = errors.New(ret.ErrMsg)
		return
	}

	return
}
