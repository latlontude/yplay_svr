package submit

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/add":                    auth.Apify2(doSubmitQuestion),             //用户投稿问题
		"/approve":                auth.Apify(doSubmitApprove),               //用户投稿审核通过
		"/batchapprove":           auth.Apify(doBatchSubmitApprove),          //用户投稿批量审核通过
		"/reject":                 auth.Apify(doSubmitReject),                //用户投稿审核拒绝
		"/update":                 auth.Apify2(doSubmitUpdate),               //用户投稿更新
		"/delete":                 auth.Apify2(doSubmitDelete),               //用户投稿审核未通过删除
		"/querylist":              auth.Apify2(doSubmitQueryList),            //用户投稿列表查询 支持审核中/审核通过/审核未通过状态
		"/querylistnotonline":     auth.Apify2(doSubmitQueryListNotOnline),   //用户未上线的投稿列表查询 包括审核中和审核未通过的
		"/querydetail":            auth.Apify2(doSubmitQueryDetail),          //用户投稿详情查询
		"/querynewonlinecnt":      auth.Apify2(doSubmitGetNewOnlineCnt),      //用户投稿详情查询
		"/querynewlyaddedhotflag": auth.Apify2(doSubmitGetNewlyAddedHotFlag), //查询用户投稿的题目有没有新增热度
	}

	log = env.NewLogger("vote")
)
