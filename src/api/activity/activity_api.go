package activity

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getmyactivityinfo": auth.Apify2(doGetMyActivityInfo), //拉取好友混排消息列表
	}

	log = env.NewLogger("activity")
)
