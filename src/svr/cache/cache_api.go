package cache

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/cachereload": auth.Apify(doCacheReload), //重新加载cache
	}

	log = env.NewLogger("cache")
)
