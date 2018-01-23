package feed

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getfeeds": auth.Apify2(doGetFeeds), //拉取好友混排消息列表 只返回比当前时间戳小的并且未读的消息
		"/ackfeeds": auth.Apify2(doAckFeeds), //确认已经收到的feeds, 服务器会删除这些feeds
	}

	log = env.NewLogger("feed")
)
