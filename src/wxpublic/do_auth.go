package wxpublic

import (
	//"crypto/sha1"
	"encoding/xml"
	"io/ioutil"
	"net/http"
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
	replyMsg.Content = "已经收到您的消息: " + recvMsg.Content
	replyMsg.CreateTime = int(time.Now().Unix())
	replyMsg.ToUserName = recvMsg.FromUserName
	replyMsg.FromUserName = recvMsg.ToUserName
	replyMsg.MsgType = "text"

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
