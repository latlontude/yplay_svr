package helper

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"regexp"
)

var (
	APIMap = httputil.APIMap{
		"/downloadredirect": auth.Apify3(doDownloadRedirect), //好友短信邀请链接地址重定向
		"/ping":             auth.Apify(doPing),
	}

	//Apple
	REG_IOS = regexp.MustCompile(`\(iPhone(.*?)\) AppleWebKit`)
	REG_MAC = regexp.MustCompile(`\(Macintosh(.*?)\) AppleWebKit`)

	//小米
	REG_MI    = regexp.MustCompile(`\(Linux(.*?); MI (.*?)Build/(.*?)\) AppleWebKit`)
	REG_REDMI = regexp.MustCompile(`\(Linux(.*?); Redmi (.*?)Build/(.*?)\) AppleWebKit`)

	//华为
	REG_HUAWEI = regexp.MustCompile(`\(Linux(.*?)HUAWEI(.*?)\) AppleWebKit`)
	REG_HONOR  = regexp.MustCompile(`\(Linux(.*?)HONOR(.*?)\) AppleWebKit`)

	//oppo/vivo
	REG_OPPO = regexp.MustCompile(`\(Linux(.*?); OPPO (.*?)Build/(.*?)\) AppleWebKit`)
	REG_VIVO = regexp.MustCompile(`\(Linux(.*?); vivo (.*?)Build/(.*?)\) AppleWebKit`)

	//三星
	REG_SM = regexp.MustCompile(`\(Linux(.*?); SM-(.*?)Build/(.*?)\) AppleWebKit`)

	//安卓
	REG_AND = regexp.MustCompile(`\(Linux(.*?)Android(.*)\) AppleWebKit`)

	log = env.NewLogger("helper")
)
