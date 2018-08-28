package board

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"svr/st"
)

var (
	APIMap = httputil.APIMap{
		"/getboards":      auth.Apify2(doGetBoards),
		"/follow":         auth.Apify2(doFollow),
		"/unfollow":       auth.Apify2(doUnfollow),
		"/join":           auth.Apify2(doJoin),         //加入天使团
		"/getAngelList":   auth.Apify2(doGetAngelInfo), //天使首页 获取天使团信息
		"/addAngel":       auth.Apify2(doAddAngel),
		"/demiseBigAngel": auth.Apify2(doDemiseBigAngel), //转让
		"/deleteAngel":    auth.Apify2(doDelAngel),       //卸任
		"/inviteAngel":    auth.Apify2(doInviteAngel),    //邀请
		"/acceptInvite":   auth.Apify2(doAcceptAngel),    //接收
	}

	log = env.NewLogger("board")

	boardMap = make(map[int]*st.BoardInfo) //boardId => boardInfo
)
