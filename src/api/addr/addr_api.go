package addr

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/update":       auth.Apify2(doUpdateAddr),   //拉取好友混排消息列表
		"/remove":       auth.Apify2(doRemoveAddr),   //拉取好友混排消息列表
		"/querybyphone": auth.Apify2(doQueryByPhone), //通过手机号查询用户信息
	}

	log = env.NewLogger("addr")
)
