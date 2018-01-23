package geneqids

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/gene":                  auth.Apify2(doGene),                  //投票操作
		"/approvequestionupdate": auth.Apify2(doApproveQuestionUpdate), //审核通过之后的更新操作
	}

	log = env.NewLogger("geneqids")
)
