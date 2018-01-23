package push

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/wnspush": auth.Apify2(doWnsPush), //wnspush test
		//"/getwnsonlinestatus":         auth.Apify2(doGetWnsOnlineStatus),  //wnspush test
	}

	log = env.NewLogger("push")
)
