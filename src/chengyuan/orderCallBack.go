/**
	统一下单接口 生成预支付表示和orderId给客户端
 */

package chengyuan

import (
	"encoding/xml"
	"net/http"
)

type CallBackResult struct {
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
	CodeUrl    string   `xml:"code_url"`
}



func OrderCallBack(req *CallBackResult, r *http.Request) (replyStr *string, err error) {
	log.Debugf("r:%v", r)
	return
}
