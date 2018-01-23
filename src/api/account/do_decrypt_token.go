package account

import (
	"common/token"
	"net/http"
)

type DecryptTokenReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type DecryptTokenRsp struct {
	Uin    int64  `json:"uin"`
	Ver    int    `json:"ver"`
	Uuid   int64  `json:"uuid"`
	Ts     int    `json:"ts"`
	Device string `json:"device"`
	Os     string `json:"os"`
	AppVer string `json:"appVer"`
}

func doDecryptToken(req *DecryptTokenReq, r *http.Request) (rsp *DecryptTokenRsp, err error) {

	log.Debugf("uin %d, DecryptTokenReq %+v", req.Uin, req)

	t, err := token.DecryptToken(req.Token, req.Ver)
	if err != nil {
		log.Errorf("uin %d, DecryptTokenRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DecryptTokenRsp{t.Uin, t.Ver, t.Uuid, t.Ts, t.Device, t.Os, t.AppVer}

	log.Debugf("uin %d, DecryptTokenRsp succ, %+v", req.Uin, rsp)

	return
}
