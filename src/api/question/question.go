package question

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/post": auth.Apify2(doPostQuestion), //发表提问
		//"/modify":         auth.Apify2(doModQuestion),  //修改提问
		//"/del":            auth.Apify2(doDelQuestion),  //删除提问
		//"/getquestions":   auth.Apify2(doGetQuestions), //根据板块ID来拉取问题列表 后续可能有按用户ID来拉取问题列表
		//"/api/getanswers": auth.Apify2(doGetAnswers),   //拉取某个提问的回答列表
	}

	log = env.NewLogger("question")
)
