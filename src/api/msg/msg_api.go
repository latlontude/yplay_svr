package msg

import (
	"common/env"
	"common/httputil"
	//"common/auth"
)

var (
	APIMap = httputil.APIMap{

	//"/dogetmsgs":        auth.Apify2(doGetMsgs),         //拉取未读消息
	//"/setmsgreaded":     auth.Apify2(doSetMsgReaded),    //设置消息已读
	//"/dovotereply":      auth.Apify2(doVoteReply),       //被投票者进行回复
	//"/dovotereplyreply": auth.Apify2(doVoteReplyReply),  //投票者对回复的回复 会暴露自己的昵称和性别信息等

	}

	log = env.NewLogger("msg")
)

func Init() (err error) {
	return
}
