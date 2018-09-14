package chengyuan

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/auth":  auth.Apify4(doAuth), //发送短信
		"/login": auth.Apify(doGetCode),
		"/order": auth.Apify(doOrder),
	}

	log = env.NewLogger("chengyuan_wxpublic")

	TOKEN = "cool6d709yeejaypupusocool"
)
