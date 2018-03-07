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

	var replyMsg Msg
	replyMsg.CreateTime = int(time.Now().Unix())
	replyMsg.ToUserName = recvMsg.FromUserName
	replyMsg.FromUserName = recvMsg.ToUserName
	replyMsg.MsgType = "text"

	//输入类型非文本消息
	if recvMsg.MsgType != "text" {

		replyMsg.Content = getContent(1)
		nd, err1 := xml.Marshal(replyMsg)
		if err1 != nil {
			log.Errorf("marshal msg error " + err1.Error())
			err = nil
			return
		}

		ndstr := string(nd)
		replyStr = &ndstr

		return
	}

	//包含注册文字 提示文案
	if strings.Contains(recvMsg.Content, "注册") {

		replyMsg.Content = "在应用商店搜索“噗噗”app，下载并注册成功后，来找我领红包～（注意：记得填写真实姓名）"
		nd, err1 := xml.Marshal(replyMsg)
		if err1 != nil {
			log.Errorf("marshal msg error " + err1.Error())
			err = nil
			return
		}

		ndstr := string(nd)
		replyStr = &ndstr

		return
	}

	ok, code, phoneNum := checkUserInput(recvMsg.FromUserName, recvMsg.Content)

	log.Errorf("checkUserInput ret %d, code %d, phoneNum %s", ok, code, phoneNum)

	if !ok {
		content := getContent(code)
		if len(content) == 0 {
			log.Errorf("getContent err :%s", content)
		} else {
			replyMsg.Content = content
		}
	} else {
		if code == SENDPHONECODE {
			yes := checkRegister(phoneNum)
			if !yes {
				content := getContent(8) // 未注册
				replyMsg.Content = content
			} else {
				yes := checkUserInfo(phoneNum)
				if yes {

					hasGet, err1 := HasGetRandomRedPacket(phoneNum)
					if err1 == nil {

						//已经领过红包
						if hasGet > 0 {
							replyMsg.Content = getContent(3)
						} else {
							//没有领取，发送验证码
							phoneCode := genePhoneCode(phoneNum)
							sendPhoneCode(phoneNum, phoneCode)
							saveCode(recvMsg.FromUserName, phoneCode, phoneNum)
							replyMsg.Content = "请输入您接收到的地大女生节活动专属验证码"
						}
					} else {
						//服务器出错
						replyMsg.Content = getContent(-1)
					}

					//go SaveUserPhone2OpenId(phoneNum, replyMsg.ToUserName)

				} else { // 非地大女生
					content := getContent(7) // 此活动只针对地大女生
					replyMsg.Content = content
				}

			}
		} else if code == SENDREDPACKET {

			log.Errorf("doSendRedPacket phoneNum %s", phoneNum)

			ret := doSendRedPacket(recvMsg.FromUserName, phoneNum)
			if ret != 0 {
				content := getContent(ret)
				replyMsg.Content = content
			} else {
				content := "已经发送红包给您，请您查收"
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

func checkRegister(phoneNum string) (ret bool) {
	log.Debugf("start checkRegister phone:%s", phoneNum)
	ret = false

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err := rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select uin from profiles where phone = %s`, phoneNum)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	var uin int64
	for rows.Next() {
		rows.Scan(&uin)
	}

	if uin > 0 {
		ret = true
	}

	log.Debugf("end checkRegister phone:%s", phoneNum)
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

	sql := fmt.Sprintf(`select nickName, gender, schoolId from profiles where phone = "%s"`, phoneNum)

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

func SaveUserPhone2OpenId(phone string, openId string) {

	log.Debugf("start save phone2openId phone:%s, openId:%s", phone, openId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err := rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert ignore into redPacketPhone2OpenId values(?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(phone, openId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("start save phone2openId phone:%s, openId:%s", phone, openId)
	return
}

func doSendRedPacket(openId string, phoneNum string) (code int) {
	log.Debugf("start doSendRedPacket openId:%s, phoneNum %s", openId, phoneNum)

	hasGet, err := HasGetRandomRedPacket(phoneNum)
	if err == nil {
		if hasGet == 1 {
			return 3 // 用户已经领取过红包
		} else {
			money, err := GetRandomRedPacket(openId, phoneNum)
			if err == nil { //
				if money == 0 {
					return 4 //红包派完了
				}

				ret, _ := SendRedPacket(openId, money*100)
				if ret == 1 { //发放成功
					UpdateRedPacketReceiveRecord(openId, phoneNum)
				}
			}
		}
	} else {
		return -1
	}

	log.Debugf("end doSendRedPacket opendId:%s, code:%d", openId, code)
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
			minTs := maxTs - 3*60                                                                                                                                // 3分钟内有效
			sql := fmt.Sprintf(`select code, phone from phoneCode where openId = "%s" and ts >= %d and ts <= %d order by ts desc limit 1`, openId, minTs, maxTs) // 3分钟内code

			rows, err := inst.Query(sql)
			if err != nil {
				err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
				log.Error(err.Error())
				return false, 0, ""
			}
			defer rows.Close()

			codeIn3 := 0
			var phoneNumT string
			for rows.Next() {
				rows.Scan(&codeIn3, &phoneNumT)
			}

			log.Errorf("checkUserInput content:%s, get phone by openId %s, code %d, phone %s", content, openId, codeIn3, phoneNumT)
			phoneNum = phoneNumT

			sql = fmt.Sprintf(`select code from phoneCode where openId = "%s" order by ts desc limit 1`, openId) // 3分钟外code

			rows, err = inst.Query(sql)
			if err != nil {
				err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
				log.Error(err.Error())
				return false, 0, phoneNum
			}
			defer rows.Close()

			codeOut3 := 0
			for rows.Next() {
				rows.Scan(&codeOut3)
			}

			log.Debugf("codeIn3:%d codeOut3:%d", codeIn3, codeOut3)
			if codeIn3 == 0 && codeOut3 == 0 {
				return false, 1, phoneNum // 没有给该用户发送验证码，该用户输入了一个手机验证码
			} else if codeIn3 == 0 && codeOut3 != 0 {
				return false, 6, phoneNum //验证码失效
			} else if codeIn3 != 0 {
				codeStr := fmt.Sprintf("%d", codeIn3)
				if codeStr == content { // 验证码校验成功
					return true, SENDREDPACKET, phoneNum
				} else { //验证码校验失败
					return false, 5, phoneNum // 重新输入验证码
				}
			}
		}
	} else {

		index := strings.Index(content, "地大女生")
		lent := len(content)

		log.Errorf("content:%s length %d, index(地大女生) %d", content, lent, index)

		//地大女生13590457127 = 15个字符
		if lent != 23 || index != 0 {
			// 输入文本格式不正确
			log.Errorf("content invalid %s", content)
			return false, 1, ""
		}

		//必须是11个字符
		phoneStr := content[12:]

		_, err := strconv.Atoi(phoneStr)
		if err != nil { // 手机号包含非数字
			log.Errorf("phoneStr:%s err ", phoneStr)
			return false, 1, phoneStr
		} else {
			ok := IsValidPhone(phoneStr)
			if ok { //是一个手机号
				return true, SENDPHONECODE, phoneStr
			} else {
				return false, 1, phoneStr //输入文本格式不正确
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
		content = "输入格式错误,请按如下格式输入：“地大女生”+“手机号”（注意:没有中间的加号哦）"
	case 2:
		content = "请输入地大女生节活动专属验证码"
	case 3:
		content = "你已经领取过红包"
	case 4:
		content = "哎呦，不好意思，红包派完了"
	case 5:
		content = "地大女生节活动专属验证码错误，请重新输入验证码"
	case 6:
		content = "地大女生节活动专属验证码已失效，请重新输入:“地大女生”+“手机号”,获取新的验证码"
	case 7:
		content = "此活动只针对地大女生"
	case 8:
		content = "该手机号还没有注册【噗噗】，请先在应用商店下载并注册【噗噗】后，来找我领红包～（注意：记得填写真实姓名）"
	default:
		content = "活动太热烈！请稍后再试"
	}

	log.Debugf("code:%d, content:%s", code, content)
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
	log.Debugf("end GenePhoneCode phone:%s code:%d", phone, code)
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
