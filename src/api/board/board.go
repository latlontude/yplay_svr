package board

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"svr/st"
)

var (
	APIMap = httputil.APIMap{
		"/getboards":    auth.Apify2(doGetBoards),
		"/follow":       auth.Apify2(doFollow),
		"/unfollow":     auth.Apify2(doUnfollow),
		"/join":         auth.Apify2(doJoin),         //加入天使团
		"/getAngelList": auth.Apify2(doGetAngelInfo), //获取天使团信息
	}

	log      = env.NewLogger("board")
	boardMap = make(map[int]*st.BoardInfo) //boardId => boardInfo
)
