package like

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/post": auth.Apify2(doPostLike), //点赞
		"/del":  auth.Apify2(doDelLike),  //取消点赞
	}

	log = env.NewLogger("like")
)
