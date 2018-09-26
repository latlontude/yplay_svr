/**
统一下单接口 生成预支付表示和orderId给客户端
*/

package express

import (
	"bytes"
	"crypto/md5"
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

const (
	SECRET = "df1176a12952f19d0008facdab60edd6"
)

type OrderReq struct {
	OrderId     int64  `json:"orderId"`
	Openid      string `json:"openid"`
	SchoolId    int    `json:"schoolId"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	ParcelInfo  string `json:"parcelInfo"`
	ParcelSize  int    `json:"parcelSize"`
	SendAddr    string `json:"sendAddr"`
	ReceiveAddr string `json:"receiveAddr"`
	ArrivalTs   int    `json:"arrivalTs"`
	Fee         int    `json:"fee"`
}

type OrderCallBack struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	AppId      string   `xml:"appid"`
	MchId      string   `xml:"mch_id"`
	DeviceInfo string   `xml:"device_info"`
	NonceStr   string   `xml:"nonce_str"`
	Sign       string   `xml:"sign"`
	ResultCode string   `xml:"result_code"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
	TradeType  string   `xml:"trade_type"`
	PrepayId   string   `xml:"prepay_id"`
}

type WxOrderInfo struct {
	XMLName    xml.Name `xml:"xml"`
	AppId      string   `xml:"appid"`        // 公众账号appid String(32)
	MchId      string   `xml:"mch_id"`       // 商户号  String(32) 微信支付分配的商户号
	DeviceInfo string   `xml:"device_info"`  // 自定义参数，可以为终端设备号(门店号或收银设备ID)，PC网页或公众号内支付可以传"WEB"
	NonceStr   string   `xml:"nonce_str"`    // 随机字符串 String(32)
	Sign       string   `xml:"sign"`         // 签名  String(32)
	SignType   string   `xml:"sign_type"`    // 签名类型，默认为MD5，支持HMAC-SHA256和MD5。
	Body       string   `xml:"body"`         // 商品简单描述，该字段请按照规范传递，具体请见参数规定
	Detail     string   `xml:"detail"`       // String(6000)
	Attach     string   `xml:"attach"`       // string 127
	OutTradeNo string   `xml:"out_trade_no"` // 商户系统内部订单号，要求32个字符内，只能是数字、大小写字母_-|* 且在同一个商户号下唯一
	FreeType   string   `xml:"free_type"`    // 默认人民币：CNY，
	TotalFee   string   `xml:"total_fee"`    // 订单总金额，单位为分，

	SpBillCreateIp string `xml:"spbill_create_ip"` //APP和网页支付提交用户端ip，Native支付填调用微信支付API的机器IP
	TimeStart      string `xml:"time_start"`       //订单生成时间，格式为yyyyMMddHHmmss，如2009年12月25日9点10分10秒表示为20091225091010
	TimeExpire     string `xml:"time_expire"`      //订单失效时间，格式为yyyyMMddHHmmss，如2009年12月27日9点10分10秒表示为20091227091010
	GoodsTag       string `xml:"goods_tag"`        //订单优惠标记，使用代金券或立减优惠功能时需要的参数
	NotifyUrl      string `xml:"notify_url"`       //异步接收微信支付结果通知的回调地址，通知url必须为外网可访问的url，不能携带参数
	TradeType      string `xml:"trade_type"`       //JSAPI 公众号支付 NATIVE 扫码支付 APP APP支付

	ProductId string `xml:"product_id"` //12235413214070356458058
	LimitPay  string `xml:"limit_pay"`  //上传此参数no_credit--可限制用户不能使用信用卡支付
	Openid    string `xml:"openid"`
	SceneInfo string `xml:"scene_info"` //上报场景信息
}

type OrderRsp struct {
	AppId     string `json:"appId"`    // 公众账号appid String(32)
	NonceStr  string `json:"nonceStr"` // 随机字符串 String(32)
	SignType  string `json:"signType"` // 签名类型，默认为MD5，支持HMAC-SHA256和MD5。
	PaySign   string `json:"paySign"`
	TimeStamp string `json:"timeStamp"`
	Package   string `json:"package"`
}

func doOrder(req *OrderReq, r *http.Request) (rsp *OrderRsp, err error) {
	log.Debugf("OrderReq,%+v", req)

	fee, err1 := GetFee(req.SchoolId, req.SendAddr, req.ReceiveAddr, req.ParcelSize)
	if err1 != nil {

	}
	//统一下单
	prepayId, rsp, err := UnifiedOrder(req.OrderId, req.Openid)
	//插入数据库
	InsertOrderInfo(req, prepayId, fee)
	return
}

func UnifiedOrder(orderId int64, openid string) (PrepayId string, rsp *OrderRsp, err error) {

	//生成随机数并md5加密随机数
	rand.Seed(time.Now().UnixNano())
	md5Ctx1 := md5.New()
	md5Ctx1.Write([]byte(strconv.Itoa(rand.Intn(1000))))
	nonceStr := hex.EncodeToString(md5Ctx1.Sum(nil)) //随机数

	//所有参数放到map里
	//paramsMap := make(map[string]string,0)
	var orderInfo WxOrderInfo
	orderInfo.AppId = WXAPPID
	orderInfo.MchId = MCH_ID
	orderInfo.Attach = "支付测试"
	orderInfo.DeviceInfo = "iphone"
	orderInfo.NonceStr = nonceStr
	orderInfo.SignType = "MD5"
	orderInfo.Body = "腾讯-游戏"
	orderInfo.OutTradeNo = fmt.Sprintf("%d", orderId)
	orderInfo.TotalFee = "10"
	orderInfo.SpBillCreateIp = "119.137.55.46"
	orderInfo.NotifyUrl = "http://yplay.vivacampus.com/api/express/orderCallBack"
	orderInfo.TradeType = "JSAPI"
	orderInfo.Openid = openid
	orderInfo.Sign = GetSign(&orderInfo, SECRET)

	reqQueryStr, err := xml.MarshalIndent(&orderInfo, "  ", "    ")
	log.Debugf("req:%s", string(reqQueryStr))

	if err != nil {
		log.Errorf("err:%+v", err)
	}

	//POST数据
	resp, err := http.Post("https://api.mch.weixin.qq.com/pay/unifiedorder",
		"text/xml",
		bytes.NewBuffer(reqQueryStr))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var orderBack OrderCallBack
	err = xml.Unmarshal(body, &orderBack)
	log.Debugf("body:%s   xml:%+v", string(body), orderBack)
	if err != nil {
		log.Errorf("xml decode fail  orderback:%+v", orderBack)
		return
	}

	now := fmt.Sprintf("%d", time.Now().Unix())
	PrepayId = orderBack.PrepayId
	pkg := fmt.Sprintf("prepay_id=%s", PrepayId)

	rsp = &OrderRsp{}
	rsp.AppId = WXAPPID
	rsp.Package = pkg
	rsp.SignType = "MD5"
	rsp.NonceStr = nonceStr
	rsp.TimeStamp = now

	rsp.PaySign = CalcPaySign(WXAPPID, now, nonceStr, pkg)

	log.Debugf("rsp:%+v", rsp)
	return
}

func CalcPaySign(appid, timestamp, nonceStr, pkg string) (paysign string) {

	t1 := fmt.Sprintf("appId=%s&nonceStr=%s&package=%s&signType=%s&timeStamp=%s",
		appid, nonceStr, pkg, "MD5", timestamp)
	t2 := fmt.Sprintf("%s&key=%s", t1, "4sbBcRjiGtWMSsOt56JYR1knwkTWe5TU")

	//md5Ctx2 := md5.New()
	//md5Ctx2.Write([]byte(t2))
	//sign := hex.EncodeToString(md5Ctx2.Sum(nil))
	//paysign = strings.ToUpper(sign) //签名

	log.Debugf("t2=%s", t2)

	sign := fmt.Sprintf("%x", md5.Sum([]byte(t2)))
	paysign = strings.ToUpper(string(sign))

	return
}
