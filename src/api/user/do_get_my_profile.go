package user

import (
	"net/http"
	"svr/st"
)

type GetMyProfileReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetMyProfileRsp struct {
	Info          *st.UserProfileInfo2      `json:"info"`
	ModInfos      []*st.ProfileModQuotaInfo `json:"modInfos"`
	FriendListVer int64                     `json:"friendListVer"` //好友列表的版本号
}

func doGetMyProfile(req *GetMyProfileReq, r *http.Request) (rsp *GetMyProfileRsp, err error) {

	log.Debugf("uin %d, GetMyProfileReq %+v", req.Uin, req)

	//h := r.Header

	//log.Debugf("request header %+v", h)

	info, err := st.GetUserProfileInfo2(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetMyProfileRsp error, %s", req.Uin, err.Error())
		return
	}

	modInfos, err := st.GetUserProfileModQuotaAllInfo(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetMyProfileRsp error, %s", req.Uin, err.Error())
		return
	}

	ver, _ := st.GetFriendListVer(req.Uin)

	rsp = &GetMyProfileRsp{info, modInfos, ver}

	log.Debugf("uin %d, GetMyProfileRsp succ, %+v", req.Uin, rsp)

	return
}
