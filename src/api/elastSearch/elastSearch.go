package elastSearch

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/essearch": auth.Apify2(doSearchInterlocutionFromEs), //获取标签列表
		//"/uploadImg": auth.Apify2(doUploadImg),
	}

	log = env.NewLogger("elastSearch")
)

//删除
//curl -XDELETE http://122.152.206.97:9200/interlocution/
