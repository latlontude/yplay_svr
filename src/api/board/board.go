package board

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getboards": auth.Apify2(doGetBoards),
		"/follow":    auth.Apify2(doFollow),
		"/unfollow":  auth.Apify2(doUnfollow),
	}

	log = env.NewLogger("board")
)
