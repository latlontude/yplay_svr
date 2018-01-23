package sns

import (
	"net/http"
	"svr/st"
)

type GetReqAddFriendUinsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetReqAddFriendUinsRsp struct {
	Uins []int64 `json:"uins"`
}

func doGetReqAddFriendUins(req *GetReqAddFriendUinsReq, r *http.Request) (rsp *GetReqAddFriendUinsRsp, err error) {

	log.Errorf("uin %d, GetReqAddFriendUinsReq %+v", req.Uin, req)

	uins, err := st.GetMyInviteUins(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetReqAddFriendUinsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetReqAddFriendUinsRsp{uins}

	log.Errorf("uin %d, GetReqAddFriendUinsRsp succ, %+v", req.Uin, rsp)

	return
}
