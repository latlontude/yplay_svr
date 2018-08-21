package board

import (
	"net/http"
)

type DelAngelReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int   `schema:"boardId"`
	AngelUin int64 `schema:"angelUin"`
}

type DelAngelRsp struct {
	Code int `json:"code"`
}

func doDelAngel(req *DelAngelReq, r *http.Request) (rsp *DelAngelRsp, err error) {

	log.Debugf("uin %d, AddAngelReq succ, %+v", req.Uin, rsp)

	err = DelAngelFromAdmin(req.Uin, req.BoardId, req.AngelUin)

	if err != nil {
		log.Errorf("uin %d, AddAngelRsp error, %s", req.Uin, err.Error())
		return
	}

	code := 0
	rsp = &DelAngelRsp{code}
	return
}
