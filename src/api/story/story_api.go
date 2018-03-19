package story

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getfriendstories":       auth.Apify2(doGetFriendStories),       //拉取好友混排story列表 只返回比当前时间戳小的并且未读的消息 24小时有效
		"/ackstories":             auth.Apify2(doAckStories),             //确认已经收到的story, 服务器会删除这些story
		"/getmystories":           auth.Apify2(doGetMyStories),           //拉取我的story列表
		"/addstory":               auth.Apify2(doAddStory),               //增加story列表
		"/delstory":               auth.Apify2(doDelStory),               //删除story列表
		"/updatestoryviewrecord":  auth.Apify2(doUpdateStoryViewRecord),  //更新动态观看记录
		"/getstoryviewrecord":     auth.Apify2(doGetStoryViewRecord),     //获取动态观看记录
		"/getstoryvideouploadsig": auth.Apify2(doGetStoryVideoUploadSig), //获取视频上传的签名
	}

	log = env.NewLogger("story")
)
