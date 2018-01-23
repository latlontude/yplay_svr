package account

import (
	"common/constant"
	"common/myredis"
	"fmt"
	"net/http"
)

type LogoutReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type LogoutRsp struct {
}

func doLogout(req *LogoutReq, r *http.Request) (rsp *LogoutRsp, err error) {

	log.Debugf("uin %d, LogoutReq %+v", req.Uin, req)

	err = Logout(req.Uin)
	if err != nil {
		log.Errorf("uin %d, LogoutRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &LogoutRsp{}

	log.Debugf("uin %d, LogoutRsp succ, %+v", req.Uin, rsp)

	return
}

func Logout(uin int64) (err error) {

	if uin == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_TOKEN)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)
	err = app.Del(keyStr) //code 缓存63秒
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
