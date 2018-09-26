/**
统一下单接口 生成预支付表示和orderId给客户端
*/

package express

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type WXOrderCallBack struct {
	XMLName  xml.Name `xml:"xml"`
	AppId    string   `xml:"appid"`
	Attach   string   `xml:"attach"`
	BankType string   `xml:"bank_type"`
	CashFee  string   `xml:"cash_fee"`

	DeviceInfo    string `xml:"device_info"`
	CashFeeType   string `xml:"cash_fee_type"`
	FeeType       string `xml:"fee_type"`
	IsSubscribe   string `xml:"is_subscribe"`
	MchId         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Openid        string `xml:"openid"`
	OutTradeNo    string `xml:"out_trade_no"`
	ResultCode    string `xml:"result_code"`
	ReturnCode    string `xml:"return_code"`
	SignType      string `xml:"sign_type"`
	Sign          string `xml:"sign"`
	TimeEnd       string `xml:"time_end"`
	TotalFee      string `xml:"total_fee"`
	TradeType     string `xml:"trade_type"`
	TransactionId string `xml:"transaction_id"`
	ReturnMsg     string `xml:"return_msg"`
	ErrCode       string `xml:"err_code"`
	ErrCodeDes    string `xml:"err_code_des"`
}

type OrderCallBackRsp struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
}

func doOrderCallBack(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	var orderCallBack WXOrderCallBack
	err = xml.Unmarshal(body, &orderCallBack)
	log.Debugf("body:%s   xml:%+v", string(body), orderCallBack)
	if err != nil {
		log.Errorf("xml decode fail  orderback:%+v", orderCallBack)
		return
	}

	var orderCallBackRsp OrderCallBackRsp

	if orderCallBack.ReturnCode == "SUCCESS" {
		sign := GetSign(orderCallBack, SECRET)
		log.Debugf("sign:%s,backSign:%s", sign, orderCallBack.Sign)
		if sign == orderCallBack.Sign {
			orderCallBackRsp.ReturnCode = "SUCCESS"
			orderCallBackRsp.ReturnMsg = "OK"
		} else {
			orderCallBackRsp.ReturnCode = "FAIL"
			orderCallBackRsp.ReturnCode = "SIGN_ERROR"
		}
	} else {
		orderCallBackRsp.ReturnCode = orderCallBack.ReturnCode
		orderCallBackRsp.ReturnMsg = "FAIL"
	}

	reqQueryStr, err := xml.MarshalIndent(&orderCallBackRsp, "  ", "    ")
	log.Debugf("xmlrsp:%s", string(reqQueryStr))
	if err != nil {
		log.Errorf("err:%+v", err)
	}

	w.Write([]byte(reqQueryStr))
	return
}
