package wxpublic

import (
	//"crypto/sha1"
	"common/constant"
	"common/mydb"
	"common/rest"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AuthReq struct {
	Sig     string `schema:"signature"`
	Ts      string `schema:"timestamp"`
	Nonce   string `schema:"nonce"`
	Echostr string `schema:"echostr"`
}

type Msg struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int      `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgId        int      `xml:"MsgId"`
}

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

const (
	SENDPHONECODE = 99
	SENDREDPACKET = 100
)

func doAuth(req *AuthReq, r *http.Request) (replyStr *string, err error) {

	log.Debugf("AuthReq %+v", req)

	/*  //校验signature
	str := fmt.Sprintf("%s%d%s", req.Nonce, req.Ts, TOKEN)

	h := sha1.New()

	io.WriteString(h, str)
	sig = fmt.Sprintf("%x", h.Sum(nil))

	if req.Sig != sig{

	}
	*/

	defer r.Body.Close()

	d, err := ioutil.ReadAll(r.Body) //获取post的数据
	if err != nil {
		err = nil
		log.Errorf("read body error " + err.Error())
		return
	}

	log.Debugf("body %+v", string(d))

	var recvMsg Msg

	err = xml.Unmarshal(d, &recvMsg)
	if err != nil {
		err = nil
		log.Errorf("xml Unmarshal error " + err.Error())
		return
	}

	log.Debugf("recv msg body %+v", recvMsg)

	if recvMsg.MsgType != "text" {
		err = nil
		log.Errorf("msg type not text ")
		return
	}

	var replyMsg Msg
	replyMsg.CreateTime = int(time.Now().Unix())
	replyMsg.ToUserName = recvMsg.FromUserName
	replyMsg.FromUserName = recvMsg.ToUserName
	replyMsg.MsgType = "text"

	ok, code, phoneNum := checkUserInput(recvMsg.Content, recvMsg.FromUserName)
	if !ok {
		content := getContent(code)
		if len(content) == 0 {
			log.Errorf("getContent err :%s", content)
		} else {
			replyMsg.Content = content
		}
	} else {
		if code == SENDPHONECODE {
			phoneCode := genePhoneCode(phoneNum)
			sendPhoneCode(phoneNum, phoneCode)
			saveCode(recvMsg.FromUserName, phoneCode, phoneNum)
			replyMsg.Content = "请输入您接收到的地大女生节活动专属验证码"
		} else if code == SENDREDPACKET {
			yes := checkUserInfo(phoneNum)
			if yes {
				ret := doSendRedPacket(recvMsg.FromUserName)
				if ret != 0 {
					content := getContent(ret)
					replyMsg.Content = content
				}
			} else {
				content := getContent(7) // 此活动只针对地大女生
				replyMsg.Content = content
			}
		}
	}

	nd, err := xml.Marshal(replyMsg)
	if err != nil {
		log.Errorf("marshal msg error " + err.Error())
		err = nil
		return
	}

	log.Debugf("reply msg body %+v", replyMsg)

	ndstr := string(nd)

	replyStr = &ndstr

	log.Debugf("reply str %+v", *replyStr)

	return
}

func checkUserInfo(phoneNum string) (ret bool) {
	log.Debugf("start checkUserInfo phone:%s", phoneNum)
	ret = false
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err := rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select nickName, gender, schoolId from profiles where phone = %s`, phoneNum)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	nickName := ""
	gender := 0
	schoolId := 0
	for rows.Next() {
		rows.Scan(&nickName, &gender, &schoolId)
	}

	if len(nickName) > 0 && gender == 2 && schoolId == 78629 {
		ret = true
	}

	log.Debugf("end checkUserInfo phone:%s", phoneNum)
	return
}
func saveCode(openId string, code int, phone string) {

	log.Debugf("start saveCode openId:%s, code:%d, phone:%s", openId, code, phone)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err := rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into phoneCode values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()
	_, err = stmt.Exec(0, openId, code, phone, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end saveCode openId:%s, code:%d, phone:%s", openId, code, phone)
	return
}

func doSendRedPacket(openId string) (code int) {
	log.Debugf("start doSendRedPacket openId:%d", openId)

	hasGet, err := HasGetRandomRedPacket(openId)
	if err != nil {
		if hasGet == 1 {
			return 3 // 用户已经领取过红包
		} else {
			money, err := GetRandomRedPacket(openId)
			if err != nil {
				if money == 0 {
					return 4 //红包派完了
				}

				ret, _ := SendRedPacket(openId, money)
				if ret == 1 { //发放成功
					UpdateRedPacketReceiveRecord(openId)
				}
			}
		}
	}

	log.Debugf("end doSendRedPacket opendId:%d, code:%d", openId, code)
	return
}

func checkUserInput(openId, content string) (ok bool, code int, phoneNum string) {
	log.Debugf("start checkUserInput openId:%s, content:%s", openId, content)

	if len(content) == 4 {
		_, err := strconv.Atoi(content)
		if err != nil { // 验证码包含非数字
			log.Errorf("can not recognition content:%s", content)
			return false, 1, ""
		} else {

			inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
			if inst == nil {
				err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
				log.Errorf(err.Error())
				return false, 1, ""
			}

			maxTs := time.Now().Unix()
			minTs := maxTs - 3*60                                                                                                                       // 3分钟内有效
			sql := fmt.Sprintf(`select code from phoneCode where openId = %s and ts >= %d and ts <= %d order by ts desc limit 1`, openId, minTs, maxTs) // 3分钟内code

			rows, err := inst.Query(sql)
			if err != nil {
				err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
				log.Error(err.Error())
				return false, 0, ""
			}
			defer rows.Close()

			codeIn3 := 0
			for rows.Next() {
				rows.Scan(&codeIn3)
			}

			sql = fmt.Sprintf(`select code from phoneCode where openId = %s`, openId) // 3分钟外code

			rows, err = inst.Query(sql)
			if err != nil {
				err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
				log.Error(err.Error())
				return false, 0, ""
			}
			defer rows.Close()

			codeOut3 := 0
			for rows.Next() {
				rows.Scan(&codeOut3)
			}

			if codeIn3 == 0 && codeOut3 == 0 {
				return false, 1, "" // 没有给该用户发送验证码，该用户输入了一个手机验证码
			} else if codeIn3 == 0 && codeOut3 != 0 {
				return false, 6, "" //验证码失效
			} else if codeIn3 != 0 {
				codeStr := fmt.Sprintf("%d", code)
				if codeStr == content { // 验证码校验成功
					return true, SENDREDPACKET, ""
				} else { //验证码校验失败
					return false, 5, "" // 重新输入验证码
				}
			}
		}
	} else {
		ret := strings.Split(content, "+")
		if len(ret) != 2 { // 输入文本格式不正确
			return false, 1, ""
		} else {
			if ret[0] == "地大女生" {
				_, err := strconv.Atoi(ret[1])
				if err != nil { // 手机号包含非数字
					return false, 1, ""
				} else {
					ok := IsValidPhone(ret[1])
					if ok { //是一个手机号
						return true, SENDPHONECODE, ret[1]
					} else {
						return false, 1, "" //输入文本格式不正确
					}
				}

			} else { // 输入文本格式不正确
				return false, 1, ""
			}
		}

	}
	log.Debugf("start checkUserInput openId:%s, content:%s", openId, content)
	return
}

func getContent(code int) (content string) {
	log.Debugf("code:%d", code)
	switch code {
	case 1:
		content = "输入格式错误,请按如下格式输入：地大女生+手机号 "
	case 2:
		content = "请输入地大女生节活动专属验证码"
	case 3:
		content = "你已经领取过红包"
	case 4:
		content = "哎呦，不好意思，红包派完了"
	case 5:
		content = "地大女生节活动专属验证码错误，请重新输入验证码"
	case 6:
		content = "地大女生节活动专属验证码已失效，请重新输入:地大女生+手机号,获取新的验证码"
	case 7:
		content = "此活动只针对地大女生"
	}
	log.Debugf("code:%d, content:%s", content)
	return
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

	return false
}

func genePhoneCode(phone string) (code int) {
	log.Debugf("start GenePhoneCode phone:%s", phone)
	randor := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := randor.Intn(9000) + 1000

	code = randNum
	log.Debugf("end GenePhoneCode phone:%s code:%s", phone, code)
	return
}

func sendPhoneCode(phone string, code int) (err error) {

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

	text := fmt.Sprintf("%d", code)
	req.Params = []string{"【噗噗】地大女生节活动验证码", text, "打死也不要告诉别人，3"}

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
