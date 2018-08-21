package board

import (
	"net/http"
)

type AcceptAngelReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int `schema:"boardId"`
	MsgId   int `schema:"msgId"`
}

type AcceptAngelRsp struct {
	Code int `json:"code"`
}

func doAcceptAngel(req *AcceptAngelReq, r *http.Request) (rsp *AcceptAngelRsp, err error) {

	log.Debugf("uin %d, AcceptAngelReq  : %+v", req.Uin, req)

	err = AcceptAngel(req.Uin, req.BoardId, req.MsgId)

	if err != nil {
		log.Errorf("uin %d, AcceptAngelReq error, %s", req.Uin, err.Error())
		return
	}

	code := 0
	rsp = &AcceptAngelRsp{code}
	return
}
