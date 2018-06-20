package chengyuan

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/auth": auth.Apify4(doAuth), //发送短信
	}

	log = env.NewLogger("chengyuan_wxpublic")

	TOKEN = "cool6d709yeejaypupusocool"
)
