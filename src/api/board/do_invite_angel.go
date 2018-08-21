package board

import (
	"net/http"
)

type InviteAngelReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	BoardId  int    `schema:"boardId"`
	AngelUin int64  `schema:"angelUin"`
}

type InviteAngelRsp struct {
	Code int `json:"code"`
}

func doInviteAngel(req *InviteAngelReq, r *http.Request) (rsp *InviteAngelRsp, err error) {

	log.Debugf("uin %d, AddAngelReq succ, %+v", req.Uin, rsp)

	err = InviteAngel(req.Uin, req.AngelUin, req.BoardId)

	if err != nil {
		log.Errorf("uin %d, AddAngelRsp error, %s", req.Uin, err.Error())
		return
	}

	code := 0
	rsp = &InviteAngelRsp{code}
	return
}
