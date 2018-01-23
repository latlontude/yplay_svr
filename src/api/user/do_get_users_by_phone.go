package user

import (
	"common/constant"
	"common/rest"

	"encoding/base64"
	"encoding/json"
	"net/http"
	"svr/cache"
	"svr/st"
)

type GetUsersbyPhoneReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Phones string `schema:"phones"`
}

type GetUsersbyPhoneRsp struct {
	Infos []*st.UserProfileInfo `json:"infos"`
}

func doGetUsersbyPhone(req *GetUsersbyPhoneReq, r *http.Request) (rsp *GetUsersbyPhoneRsp, err error) {

	log.Debugf("uin %d, GetUsersbyPhoneReq %+v", req.Uin, req)

	res, err := GetUsersbyPhone(req.Uin, req.Phones)
	if err != nil {
		log.Errorf("uin %d, GetUsersbyPhoneRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetUsersbyPhoneRsp{res}

	log.Debugf("uin %d, GetUsersbyPhoneRsp succ, %+v", req.Uin, rsp)

	return
}

func GetUsersbyPhone(uin int64, data string) (res []*st.UserProfileInfo, err error) {

	res = make([]*st.UserProfileInfo, 0)

	decodeData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Errorf(err.Error())
		return
	}

	var phones []string
	err = json.Unmarshal([]byte(decodeData), &phones)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Errorf(err.Error())
		return
	}

	uins := make([]int64, 0)

	for _, phone := range phones {

		if len(phone) == 0 {
			continue
		}

		if uid, ok := cache.PHONE2UIN[phone]; ok {

			if uid > 0 {
				uins = append(uins, uid)
			}
		}
	}

	r, err := st.BatchGetUserProfileInfo(uins)
	if err != nil {
		return
	}

	for _, phone := range phones {

		if uid, ok := cache.PHONE2UIN[phone]; ok {

			if v, ok1 := r[uid]; ok1 {
				res = append(res, v)
			}
		}
	}

	return
}
