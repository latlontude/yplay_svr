package sns

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getaddfriendnewmsgcnt": auth.Apify2(doGetAddFriendNewMsgCnt), //拉取加好友消息数未读列表
		"/getaddfriendmsgs":      auth.Apify2(doGetAddFriendMsg),       //拉取加好友消息数未处理的列表
		"/searchfriend":          auth.Apify2(doSearchFriend),          //拉取好友列表分类型  通讯录好友/好友的好友/同校好友

		"/getrecommends": auth.Apify2(doGetRecommends), //在子页面拉取推荐好友列表
		//"/getrecommendsall":         auth.Apify2(doGetRecommendsAll),         //在加好友页面拉取各种类型的推荐好友列表
		"/getrandomrecommends": auth.Apify2(doGetRandomRecommends),  //随机推荐2位好友
		"/getuserstatuswithme": auth.Apify2(doGetUsersStatusWithMe), //拉取好友与我当前的关系状态

		"/acceptaddfriend":    auth.Apify2(doAcceptAddFriend),    //接受加好友请求
		"/addfriend":          auth.Apify2(doAddFriend),          //发送加好友请求
		"/invitefriendsbysms": auth.Apify2(doInviteFriendsBySms), //邀请好友 通过发送SMS
		"/removefriend":       auth.Apify2(doRemoveFriend),       //解除好友关系 双向解除

		"/getreqaddfrienduins": auth.Apify2(doGetReqAddFriendUins), //已经申请加好友的ID列表
	}

	log = env.NewLogger("sns")
)

func Init() (err error) {
	return
}
