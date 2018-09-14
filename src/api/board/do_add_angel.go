package board

import (
	"net/http"
)

type AddAngelReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int   `schema:"boardId"`
	LabelId  int   `schema:"labelId"`
	AngelUin int64 `schema:"angelUin"`
}

type AddAngelRsp struct {
	Code int `json:"code"`
}

func doAddAngel(req *AddAngelReq, r *http.Request) (rsp *AddAngelRsp, err error) {

	log.Debugf("uin %d, AddAngelReq succ, %+v", req.Uin, rsp)

	err = AddAngelInAdmin(req.Uin, req.BoardId, req.LabelId, req.AngelUin,0)

	if err != nil {
		log.Errorf("uin %d, AddAngelRsp error, %s", req.Uin, err.Error())
		return
	}

	code := 0
	rsp = &AddAngelRsp{code}
	return
}
