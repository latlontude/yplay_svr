/**
	统一下单接口 生成预支付表示和orderId给客户端
 */

package chengyuan

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	SECRET = "df1176a12952f19d0008facdab60edd6"
)

type OrderReq struct {
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
	Code int `json:"code"`
}

func doOrder(r *http.Request) (rsp *OrderRsp, err error) {
	code := UnifiedOrder()
	rsp = &OrderRsp{code}
	return
}
func UnifiedOrder() (code int) {

	//生成随机数并md5加密随机数
	rand.Seed(time.Now().UnixNano())
	md5Ctx1 := md5.New()
	md5Ctx1.Write([]byte(strconv.Itoa(rand.Intn(1000))))
	nonceStr := hex.EncodeToString(md5Ctx1.Sum(nil)) //随机数

	//所有参数放到map里
	//paramsMap := make(map[string]string,0)
	var req OrderReq
	req.AppId = WXAPPID
	req.MchId = MCH_ID
	req.Attach = "支付测试"
	req.DeviceInfo = "iphone"
	req.NonceStr = nonceStr
	req.SignType = "MD5"
	req.Body = "腾讯-游戏"
	req.OutTradeNo = "201709141111"
	req.TotalFee = "10"
	req.SpBillCreateIp = "119.137.55.46"
	req.NotifyUrl = "http://ddactive.yeejay.com/api/orderCallBack"
	req.TradeType = "JSAPI"
	req.Openid = "o9X_y1TqF2xgqC8LdQ5Fyg97FxMI"
	//获得sign
	req.Sign = GetSign(&req, SECRET)

	reqQueryStr, err := xml.MarshalIndent(&req, "  ", "    ")
	log.Debugf("req:%s", string(reqQueryStr))

	if err != nil {
		log.Errorf("err:%+v", err)
	}

	log.Errorf("sss:%s", string(reqQueryStr))

	//POST数据
	resp, err := http.Post("https://api.mch.weixin.qq.com/pay/unifiedorder",
		"text/xml",
		bytes.NewBuffer(reqQueryStr))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Debugf("order body:%+v", string(body))

	return
}

func GetSign(req *OrderReq, secret string) (sign string) {

	paramsMap := make(map[string]interface{}, 0)
	object := reflect.ValueOf(req)
	ref := object.Elem()
	typeOfType := ref.Type()
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Field(i)
		tag := typeOfType.Field(i).Tag.Get("xml")
		if field.Interface() != "" && tag != "xml" {
			if field.Type().Name() == "int" && field.Int() == 0 {
				continue
			}
			paramsMap[tag] = field.Interface()
		}
	}

	//按照key字典顺序排序
	keys := make([]string, 0)
	for k, _ := range paramsMap {
		keys = append(keys, k)
	}
	log.Debugf("params:%+v\n", paramsMap)

	sort.Strings(keys)
	var stringA string
	for _, key := range keys {
		stringA += fmt.Sprintf("%s=%s&", key, paramsMap[key])
	}
	//最后拼接秘钥
	stringA += fmt.Sprintf("%s=%s", "key", "4sbBcRjiGtWMSsOt56JYR1knwkTWe5TU")
	log.Debugf("%s", stringA)

	md5Ctx2 := md5.New()
	md5Ctx2.Write([]byte(stringA))
	sign = hex.EncodeToString(md5Ctx2.Sum(nil))
	sign = strings.ToUpper(sign) //签名

	return
}
