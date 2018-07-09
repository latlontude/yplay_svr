package im

import "net/http"

type GetSessionIdReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	FriendUin int64 `schema:"friendUin"`
}

type GetSessionIdRsp struct {
	SessionId string `json:"sessionId"`
}

func doGetSessionId(req *GetSessionIdReq, r *http.Request) (rsp *GetSessionIdRsp, err error) {

	log.Debugf("uin %d, GetSessionIdReq %+v", req.Uin, req)

	sessionId, err := GetSnapSession(req.Uin, req.FriendUin)
	if err != nil {
		log.Errorf("uin %d, GetSessionIdReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetSessionIdRsp{sessionId}
	log.Debugf("uin %d, GetSessionIdRsp succ, %+v", req.Uin, rsp)

	return
}
