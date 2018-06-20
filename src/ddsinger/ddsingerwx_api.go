package ddsinger

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/auth": auth.Apify4(doAuth), //发送短信
	}

	log = env.NewLogger("ddsinger_wxpublic")

	TOKEN = "cool6d709yeejaypupusocool"
)
