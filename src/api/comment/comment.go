package comment

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/post":               auth.Apify2(doPostComment),       //发表评论
		"/delcomment":         auth.Apify2(doDelComment),        //删除评论
		"/replytocomment":     auth.Apify2(doReplyToComment),    //对评论进行回复
		"/replytoreply":       auth.Apify2(doReplyToReply),      //对评论下的回复进行回复
		"/delreply":           auth.Apify2(doDelReply),          //删除评论下的某个回复
		"/moveReplyToComment": auth.Apify(doMoveReplyToComment), //删除评论下的某个回复
	}

	log = env.NewLogger("comment")
)
