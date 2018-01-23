package user

import (
	"net/http"
	"svr/st"
)

type GetMyProfileModQuotaInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Field int `schema:"field"`
}

type GetMyProfileModQuotaInfoRsp struct {
	Info *st.ProfileModQuotaInfo `json:"info"`
}

func doGetMyProfileModQuotaInfo(req *GetMyProfileModQuotaInfoReq, r *http.Request) (rsp *GetMyProfileModQuotaInfoRsp, err error) {

	log.Debugf("uin %d, GetMyProfileModQuotaInfoReq %+v", req.Uin, req)

	info, err := st.GetUserProfileModQuotaInfo(req.Uin, req.Field)
	if err != nil {
		log.Errorf("uin %d, GetMyProfileModQuotaInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetMyProfileModQuotaInfoRsp{info}

	log.Debugf("uin %d, GetMyProfileModQuotaInfoRsp succ, %+v", req.Uin, rsp)

	return
}
