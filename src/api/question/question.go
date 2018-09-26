package question

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/post":                    auth.Apify2(doPostQuestion),            //发表提问
		"/modify":                  auth.Apify2(doModifyQuestion),          //修改提问
		"/del":                     auth.Apify2(doDelQuestion),             //删除提问
		"/poitags":                 auth.Apify2(doGetPoiTagList),           //获取Poi List
		"/getquestions":            auth.Apify2(doGetQuestions),            //根据板块ID来拉取问题列表 后续可能有按用户ID来拉取问题列表
		"/getgeoquestions":         auth.Apify2(doGetGeoQuestions),         //同getquestions，增加地理位置过滤条件
		"/getanswers":              auth.Apify2(doGetAnswers),              //拉取某个提问的回答列表
		"/getv2questionsforme":     auth.Apify2(doGetV2QuestionsForMe),     //拉取某个提问的回答列表
		"/getv2questionsforfriend": auth.Apify2(doGetV2QuestionsForFriend), //拉取某个提问的回答列表

		"/sameAsk": auth.Apify2(doSameAsk), //同问

		//TEST uidArr
		"/getSameAskUid":                auth.Apify2(doGetSameAskUid),
		"/dailyCount":                   auth.Apify(doDailyCount),
		"/autoQuestion":                 auth.Apify(doAutoQuestion),
		"/autoInput":                    auth.Apify(doAutoInput),
		"/getOneQuestion":               auth.Apify2(doGetOneQuestions), //根据板块ID来拉取问题列表 后续可能有按用户ID来拉取问题列表
		"/autoInputQuestionNoBoardInfo": auth.Apify(doAutoInputQuestionNoBoardInfo),
		"/getquestion":                  auth.Apify2(doGetOneQuestionDetail),

		"/readQuestion": auth.Apify2(doReadQuestion),
		"/getHeatMap":   auth.Apify(doGetHeatMap),
	}

	log = env.NewLogger("question")

	auto_uids = make([]int64, 0)
)
