package vote

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{

		"/dovote": auth.Apify2(doVote),         //投票操作
		"/doskip": auth.Apify2(doSkipQuestion), //跳过问题的操作

		"/getquestionandoptions": auth.Apify2(doGetQuestionAndOptions), //拉取一个问题和选项列表
		"/getoptions":            auth.Apify2(doGetOptions),            //选项列表

		"/submitquestion":    auth.Apify2(doSubmitQuestion),    //用户投稿问题
		"/queryrankinglist":  auth.Apify2(doGetRankingList),    // 获取用户正在答题的排行榜数据
		"/getstarinlastweek": auth.Apify2(doGetStarInLastWeek), // 获取本周话题明星
		"/releasefrozen":     auth.Apify(doReleaseFrozen),      // 解除冷冻
	}

	log = env.NewLogger("vote")
)
