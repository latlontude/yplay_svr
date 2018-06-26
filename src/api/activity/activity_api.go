package activity

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"strconv"
	"strings"
)

var (
	APIMap = httputil.APIMap{
		"/getmyactivityinfo": auth.Apify2(doGetMyActivityInfo), //拉取好友混排消息列表
	}

	log = env.NewLogger("activity")

	OpenSchools map[int]int
)

func Init() (err error) {

	OpenSchools = make(map[int]int)

	schools := strings.Split(env.Config.Activity.Schools, ",")

	for _, s := range schools {
		sid, _ := strconv.Atoi(s)
		if sid > 0 {
			OpenSchools[sid] = 1
		}
	}

	return
}
