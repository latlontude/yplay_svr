package chengyuan

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var _tlsConfig *tls.Config

type SendRedPacketReqData struct {
	XMLName      xml.Name `xml:"xml"`
	Nonce_str    string   `xml:"nonce_str"`    // 随机字符串 String(32)
	Sign         string   `xml:"sign"`         // 签名  String(32)
	Mch_billno   string   `xml:"mch_billno"`   // 商户订单号  String(28) 商户订单号（每个订单号必须唯一。取值范围：0~9，a~z，A~Z）接口根据商户订单号支持重入，如出现超时可再调用。
	Mch_id       string   `xml:"mch_id"`       //商户号  String(32) 微信支付分配的商户号
	Wxappid      string   `xml:"wxappid"`      //公众账号appid String(32)
	Send_name    string   `xml:"send_name"`    //商户名称 	String(32)  红包发送者名称
	Re_openid    string   `xml:"re_openid"`    // 用户openid  String(32) 接受红包的用户 //o9X_y1VSAIy95w3UA56knQpKMLwI
	Total_amount int      `xml:"total_amount"` // 付款金额 单位分
	Total_num    int      `xml:"total_num"`    //  红包发放总人数
	Wishing      string   `xml:"wishing"`      // 红包祝福语  String(128)
	Client_ip    string   `xml:"client_ip"`    // 调用接口的机器Ip地址 String(15)
	Act_name     string   `xml:"act_name"`     // 活动名称 String(32)
	Remark       string   `xml:"remark"`       // 备注信息 String(256)
}

type SendRedPacketRspData struct {
	XMLName      xml.Name `xml:"xml"`
	Return_code  string   `xml:"return_code"`  // 返回状态码 String(16)
	Return_msg   string   `xml:"return_msg"`   //返回信息 String(128)  如非空，为错误原因
	Sign         string   `xml:"sign"`         // String(32)
	Result_code  string   `xml:"result_code"`  //String(16)
	Err_code     string   `xml:"err_code"`     //
	Err_code_des string   `xml:"err_code_des"` //
	Mch_billno   string   `xml:"mch_billno"`   //
	Mch_id       string   `xml:"mch_id"`       //
	Wxappid      string   `xml:"wxappid"`      //
	Re_openid    string   `xml:"re_openid"`    //
	Total_amount int      `xml:"total_amount"` //
	Send_listid  string   `xml:""send_listid`  //
}

const (
	//ACT_NAME  = "女生节噗噗注册得红包活动"
	ACT_NAME  = "城院注册【噗噗】活动"
	CLIENT_IP = "115.159.147.142"
	MCH_ID    = "1498480632"
	REMARK    = "先到先得，赶快下载【噗噗】注册吧！"
	//SEND_NAME              = "深圳悦智网络科技有限公司"
	SEND_NAME = "噗噗"
	WISHING   = "在噗噗看你多受欢迎！"
	TOTAL_NUM = 1
	//TOTAL_AMOUNT           = 100
	WXAPPID                = "wxc6a993bffd64bb9a"
	KEY                    = "4sbBcRjiGtWMSsOt56JYR1knwkTWe5TU" //key为商户平台设置的密钥key
	WECHATSENDREDPACKETURL = "https://api.mch.weixin.qq.com/mmpaymkttransfers/sendredpack"
	WECHATCERTPATH         = "/home/work/yplay_svr/etc/apiclient_cert.pem"
	WECHATKEYPATH          = "/home/work/yplay_svr/etc/apiclient_key.pem"
	WECHATCAPATH           = "/home/work/yplay_svr/etc/rootca.pem"
)

type SendRedPacketReq struct {
	OpenId string `schema:"openId"`
	Money  int    `schema:"money"`
}

type SendRedPacketRsp struct {
}

/*func doSendRedPacket(req *SendRedPacketReq, r *http.Request) (rsp *SendRedPacketRsp, err error) {

	log.Errorf(" start doSendRedPacket")

	code, err := SendRedPacket(req.OpenId, req.Money)
	if err != nil {
		log.Errorf("sendRedPacket error, %s", err.Error())
		return
	}

	log.Errorf("doSendRedPacketRsp succ, %+v", rsp)

	return
}
*/
func SendRedPacket(openId string, money int) (code int, err error) {
	log.Debugf("start SendRedPacket openId:%s", openId)
	var data SendRedPacketReqData

	//生成随机数并md5加密随机数
	rand.Seed(time.Now().UnixNano())
	md5Ctx1 := md5.New()
	md5Ctx1.Write([]byte(strconv.Itoa(rand.Intn(1000))))
	nonce_str := hex.EncodeToString(md5Ctx1.Sum(nil)) //随机数

	//订单号
	//mch_billno := MCH_ID + time.Now().Format("20060102") + strconv.FormatInt(time.Now().Unix(), 10)

	mch_billno := MCH_ID + fmt.Sprintf("00000%d", time.Now().UnixNano()/1000000)

	// 生成签名
	s1 := "act_name=" + ACT_NAME + "&client_ip=" + CLIENT_IP + "&mch_billno=" + mch_billno + "&mch_id=" + MCH_ID + "&nonce_str=" + nonce_str + "&re_openid=" + openId + "&remark=" + REMARK + "&send_name=" + SEND_NAME + "&total_amount=" + strconv.Itoa(money) + "&total_num=" + strconv.Itoa(TOTAL_NUM) + "&wishing=" + WISHING + "&wxappid=" + WXAPPID + "&key=" + KEY
	md5Ctx2 := md5.New()
	md5Ctx2.Write([]byte(s1))
	s1 = hex.EncodeToString(md5Ctx2.Sum(nil))
	sign := strings.ToUpper(s1) //签名

	data.Nonce_str = nonce_str
	data.Sign = sign
	data.Mch_billno = mch_billno
	data.Mch_id = MCH_ID
	data.Wxappid = WXAPPID
	data.Send_name = SEND_NAME
	data.Re_openid = openId   //用户opendid
	data.Total_amount = money // money  单位为分
	data.Total_num = TOTAL_NUM
	data.Wishing = WISHING
	data.Client_ip = CLIENT_IP
	data.Act_name = ACT_NAME
	data.Remark = REMARK

	log.Debugf("data:%+v", data)
	output, err := xml.MarshalIndent(&data, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	log.Debugf("output:%s", string(output))

	//POST数据
	code, err = SecurePost(WECHATSENDREDPACKETURL, output)
	if err != nil {
		log.Errorf("SendRedPacketRsp err : %s", err)
	}

	log.Debugf("end SendRedPacket code:%+v", code)
	return
}

func getTLSConfig() (*tls.Config, error) {
	log.Debugf("start getTLSConfig")
	if _tlsConfig != nil {
		return _tlsConfig, nil
	}

	// load cert
	cert, err := tls.LoadX509KeyPair(WECHATCERTPATH, WECHATKEYPATH)
	if err != nil {
		log.Errorf("load wechat keys fail :%s", err)
		return nil, err
	}

	// load root ca
	caData, err := ioutil.ReadFile(WECHATCAPATH)
	if err != nil {
		log.Errorf("read wechat ca fail :%s", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	_tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	log.Debugf("end getTLSConfig")
	return _tlsConfig, nil
}

func SecurePost(url string, xmlContent []byte) (code int, err error) {
	tlsConfig, err := getTLSConfig()
	if err != nil {
		log.Debugf("getTLSConfig err :%s", err.Error())
		return 0, err
	}

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}

	rsp, err := client.Post(
		url,
		"text/xml",
		bytes.NewBuffer(xmlContent))
	if err != nil {
		log.Errorf("err in send :%s", err.Error())
	} else {

		defer rsp.Body.Close()

		if rsp.StatusCode != 200 {
			log.Errorf("rep.StatusCode != 200")
			return
		}

		body, err1 := ioutil.ReadAll(rsp.Body)
		if err1 != nil {
			log.Errorf("ioutil.ReadAll err :%s", err1.Error())
			return
		}

		var ret SendRedPacketRspData
		err1 = xml.Unmarshal(body, &ret)
		if err1 != nil {
			log.Errorf("xml.Unmarshal err :%s", err1.Error())
			return
		}

		if ret.Return_code == "SUCCESS" && ret.Result_code == "SUCCESS" {
			code = 1
		}

		log.Debugf("rsp:%+v", ret)
	}

	return
}
