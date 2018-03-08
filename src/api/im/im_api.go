package im

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/geneusersig":                       auth.Apify2(doGeneUserSig),         //生成IM消息签名
		"/syncaccount":                       auth.Apify2(doSyncAccount),         //生成IM消息签名
		"/sendvotemsg":                       auth.Apify2(doSendVoteMsg),         //第一次投票
		"/sendvotereplymsg":                  auth.Apify2(doSendVoteReplyMsg),    //投票的第一次回复
		"/sendaddfriendmsg":                  auth.Apify2(doSendAddFriendMsg),    //添加好友的push走IM通道
		"/sendleavefrozenmsg":                auth.Apify2(doSendLeaveFrozenMsg),  //冷却解除的push走IM通道
		"/sendnewfeedmsg":                    auth.Apify2(doSendNewFeedMsg),      //有新动态的push走IM通道
		"/sendremovefriendmsg":               auth.Apify2(doSendRemoveFriendMsg), //移除好友，通知被移除的一方
		"/sendlogsettingmsg":                 auth.Apify2(doSendLogSettingMsg),   //通知用户设置日志级别或者上传日志
		"/creategroup":                       auth.Apify2(doCreateGroup),
		"/sendsubmitquestionapprovedmsg":     auth.Apify2(doSendSubmitQustionApprovedMsg), //移除好友，通知被移除的一方
		"/sendsubmitnewlyaddedhotnotifymsg":  auth.Apify2(doSendSubmitVotedNotifyMsg),     //
		"/createsnapchatsession":             auth.Apify2(doCreateSnapChatSession),
		"/sendvotereplyreplymsg":             auth.Apify2(doSendVoteReplyReplyMsg),             //回复的回复
		"/batchgetsnapsessionsfroupgradeapp": auth.Apify2(doBatchGetSnapSessionsForUpgradeApp), //APP升级 批量拉取snapsessions
	}

	IM_SIG_ADMIN         string
	IM_SIG_ADMIN_GENE_TS int

	log = env.NewLogger("im")

	ChanFeedPush = make(chan int64, 10000)

	Base64PriKeyString = `LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tDQpNSUdIQWdFQU1CTUdCeXFHU000OUFnRUdDQ3FHU000OUF3RUhCRzB3YXdJQkFRUWdIODRPa2tlSHhUTnNnZXJQDQpTazRyQlphTFVHUlJ4YmRrdnNSbk5HeW1IRStoUkFOQ0FBUlg3TCtIOWhhRFQ3Wmw2MUhGcUlCRmZPR0JncCsrDQpDbDNkdlRYemVIY0JoR3VpRGJzSCtndVlRWkJjMmVnRzhmVHZudGcxUEp6UjhHSUNnT2UrVGpBWQ0KLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQ==`
)

func Init() (err error) {

	go NewFeedPushRoutine()

	return
}
