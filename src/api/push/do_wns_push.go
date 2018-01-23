package push

import (
	"bytes"
	"common/constant"
	"common/rest"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type WnsPushReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Title   string `schema:"title"`
	Content string `schema:"content"`
	Ext     string `schema:"ext"`
}

type WnsPushRsp struct {
	WId int64 `json:"wid"`
}

type WnsPushRspInfo struct {
	ErrNo  int           `json:"errno"`
	Detail []*DetailInfo `json:"detail"`
}

type DetailInfo struct {
	Ret int   `json:"ret"`
	Wid int64 `json:"wid"`
}

func doWnsPush(req *WnsPushReq, r *http.Request) (rsp *WnsPushRsp, err error) {

	log.Debugf("uin %d, WnsPushReq %+v", req.Uin, req)

	// h := r.Header

	// log.Errorf("request header %+v", h)

	// wid := ""
	// if v, ok := h["X-Wns-Wid"]; ok{
	// 	wid = strings.Join(v,";")
	// }

	wid, err := WnsPush(req.Uin, req.Title, req.Content, req.Ext)
	if err != nil {
		log.Errorf("uin %d, WnsPushRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &WnsPushRsp{wid}

	log.Debugf("uin %d, WnsPushRsp succ, %+v", req.Uin, rsp)

	return
}

func WnsPush(uin int64, title, content string, ext string) (wid int64, err error) {

	if uin == 0 || len(content) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	APPID := 203682

	SECRETID := "AKIDTq58zeASceUDl88IARl087uVu7C6PL1V"
	SECRETKEY := "rfHgtAQUWRNCuVTL5hGigcQ6e4NYbaO1"

	now := int(time.Now().Unix())

	tag := "YPLAYAHA"

	src := fmt.Sprintf("%d&%d", APPID, now)

	mac := hmac.New(sha1.New, []byte(SECRETKEY))
	mac.Write([]byte(src))
	hash := mac.Sum(nil)
	src2 := fmt.Sprintf("%s", string(hash))
	sign := base64.StdEncoding.EncodeToString([]byte(src2))

	escapedSign := url.QueryEscape(sign)

	aps := fmt.Sprintf(`{"aps":{"alert":{"title":"%s","body":"%s"},"sound":""}}`, title, content)

	params := fmt.Sprintf(`appid=%d&secretid=%s&sign=%s&tm=%d&uid=%d&tag=%s&content=%s&aps=%s`, APPID, SECRETID, escapedSign, now, uin, tag, content, aps)

	log.Errorf("WnsPush params %s", params)

	//url := fmt.Sprintf("http://wns.api.qcloud.com/api/send_msg_new")
	url := fmt.Sprintf("http://wnslog.qcloud.com/api/send_msg_new")

	hrsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer([]byte(params)))
	if err != nil {
		err = rest.NewAPIError(constant.E_PUSH_WNS_TEST, err.Error())
		log.Error(err.Error())
		return
	}

	body, err := ioutil.ReadAll(hrsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_PUSH_WNS_TEST, err.Error())
		log.Error(err.Error())
		return
	}

	log.Errorf("WnsPushRspBody %+v", string(body))

	var rsp WnsPushRspInfo
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_PUSH_WNS_TEST, err.Error())
		log.Error(err.Error())
		return
	}

	log.Errorf("WnsPushRsp %+v", rsp)

	if rsp.ErrNo != 0 {
		err = rest.NewAPIError(constant.E_PUSH_WNS_TEST, "push error")
		log.Error(err.Error())
		return
	}

	if len(rsp.Detail) == 0 {
		err = rest.NewAPIError(constant.E_PUSH_WNS_TEST, "no wid return")
		log.Error(err.Error())
		return
	}

	wid = rsp.Detail[0].Wid

	return
}
