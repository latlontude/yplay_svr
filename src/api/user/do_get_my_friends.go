package user

import (
	"net/http"
	"svr/st"
)

type GetMyFriendsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetMyFriendsRsp struct {
	Total         int              `json:"total"`
	Friends       []*st.FriendInfo `json:"friends"`
	FriendListVer int64            `json:"friendListVer"`
}

func doGetMyFriends(req *GetMyFriendsReq, r *http.Request) (rsp *GetMyFriendsRsp, err error) {

	log.Debugf("uin %d, GetMyFriendsReq %+v", req.Uin, req)

	total, friends, err := st.GetMyFriends(req.Uin, req.PageNum, req.PageSize)
	if err != nil {
		log.Errorf("uin %d, GetMyFriendsRsp error, %s", req.Uin, err.Error())
		return
	}

	ver, _ := st.GetFriendListVer(req.Uin)

	rsp = &GetMyFriendsRsp{total, friends, ver}

	log.Debugf("uin %d, GetMyFriendsRsp succ, %+v", req.Uin, rsp)

	return
}
