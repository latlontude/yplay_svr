package answer

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/post":        auth.Apify2(doPostAnswer),  //发表回答
		"/del":         auth.Apify2(doDelAnswer),   //删除回答
		"/getcomments": auth.Apify2(doGetComments), //拉取某个回答的评论列表
	}

	log = env.NewLogger("answer")
)
