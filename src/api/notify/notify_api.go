package notify

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getnewnotifystat": auth.Apify2(doGetNewNotifyStat), //wnspush test
	}

	log = env.NewLogger("notify")
)
