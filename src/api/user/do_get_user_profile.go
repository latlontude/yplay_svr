package user

import (
	"common/constant"
	"net/http"
	"svr/st"
)

type GetUserProfileReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	UserUin int64 `schema:"userUin"`
}

type GetUserProfileRsp struct {
	Info   *st.UserProfileInfo2 `json:"info"`   //用户资料
	Status int                  `json:"status"` //我跟查询者之间的状态
}

func doGetUserProfile(req *GetUserProfileReq, r *http.Request) (rsp *GetUserProfileRsp, err error) {

	log.Debugf("uin %d, GetUserProfileReq %+v", req.Uin, req)

	info, err := st.GetUserProfileInfo2(req.UserUin)
	if err != nil {
		log.Errorf("uin %d, GetUserProfileRsp error, %s", req.Uin, err.Error())
		return
	}

	ret, err := st.IsFriend(req.Uin, req.UserUin)
	if err != nil {
		log.Errorf("uin %d, GetUserProfileRsp error, %s", req.Uin, err.Error())
		return
	}

	status := constant.ENUM_SNS_STATUS_NOT_FRIEND

	if ret > 0 {

		status = constant.ENUM_SNS_STATUS_IS_FRIEND

	} else {

		ret, err = st.CheckIsMyInvite(req.Uin, req.UserUin)
		if err != nil {
			log.Errorf("uin %d, GetUserProfileRsp error, %s", req.Uin, err.Error())
			return
		}

		if ret > 0 {
			status = constant.ENUM_SNS_STATUS_HAS_INVAITE_FRIEND
		}
	}

	rsp = &GetUserProfileRsp{info, status}

	log.Debugf("uin %d, GetUserProfileRsp succ, %+v", req.Uin, rsp)

	return
}
