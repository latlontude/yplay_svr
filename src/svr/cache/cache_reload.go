package cache

import (
	"common/constant"
	"common/rest"
	"net/http"
)

type CacheReloadReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Type int `schema:"type"`
}

type CacheReloadRsp struct {
}

func doCacheReload(req *CacheReloadReq, r *http.Request) (rsp *CacheReloadRsp, err error) {

	log.Debugf("CacheReloadReq %+v", req)

	err = CacheReload(req.Uin, req.Type)
	if err != nil {
		return
	}

	rsp = &CacheReloadRsp{}

	log.Debugf("CacheReloadRsp %+v", rsp)

	return
}

func CacheReload(uin int64, typ int) (err error) {

	if uin == 0 || typ <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	if typ == 1 {
		err = CacheQuestions()
	}

	if typ == 2 {
		err = CacheSchools()
	}

	if typ == 3 {
		err = CachePhones()
	}

	if typ == 4 {
		err = CacheQIcons()
	}

	return
}
