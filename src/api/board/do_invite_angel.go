package board

import (
	"net/http"
	"strconv"
	"strings"
)

type InviteAngelReq struct {
	Uin        int64  `schema:"uin"`
	Token      string `schema:"token"`
	Ver        int    `schema:"ver"`
	BoardId    int    `schema:"boardId"`
	InviteUins string `schema:"inviteUins"`
}

type InviteAngelRsp struct {
	Code int `json:"code"`
}

func doInviteAngel(req *InviteAngelReq, r *http.Request) (rsp *InviteAngelRsp, err error) {

	uinListString := strings.Trim(req.InviteUins, ",")
	uinStrList := strings.Split(uinListString, ",")
	log.Debugf("uin %d, InviteAngelReq succ, %+v,uinStrList:%s", req.Uin, req, uinStrList)

	for _, uidStr := range uinStrList {
		uid, err1 := strconv.ParseInt(uidStr, 10, 64)
		if err1 != nil {
			log.Debugf("string to int64 error,uid:%d", uid)
			continue
		}

		err = InviteAngel(req.Uin, uid, req.BoardId)
		if err != nil {
			log.Errorf("uin %d, InviteAngelRsp error, %s", req.Uin, err.Error())
			return
		}
	}

	code := 0
	rsp = &InviteAngelRsp{code}
	return
}
