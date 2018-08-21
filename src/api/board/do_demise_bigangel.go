package board

import (
	"net/http"
)

type DemiseBigAngelReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int   `schema:"boardId"`
	AngelUin int64 `schema:"angelUin"`
}

type DemiseBigAngelRsp struct {
	Code int `json:"code"`
}

func doDemiseBigAngel(req *DemiseBigAngelReq, r *http.Request) (rsp *DemiseBigAngelRsp, err error) {

	log.Debugf("uin %d, DemiseBigAngelReq succ, %+v", req.Uin, req)

	err = DemiseBigAngel(req.Uin, req.AngelUin, req.BoardId)

	if err != nil {
		log.Errorf("uin %d, DemiseBigAngelReq error, %s", req.Uin, err.Error())
		return
	}

	code := 0
	rsp = &DemiseBigAngelRsp{code}
	return
}
