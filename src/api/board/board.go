package board

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getboards": auth.Apify2(doGetBoards), //投票操作
	}

	log = env.NewLogger("board")
)
