package im

import (
	"common/env"
	"time"
)

func NewFeedPushRoutine() {

	users := make(map[int64]int)

	for {
		select {
		case uin := <-ChanFeedPush:
			users[uin] = 1

			//独立用户累计到100, 然后集中发一次，起到合并的左右
			if len(users) > env.Config.Feed.PushMergeUserCnt {
				go SendNewFeedPush(users)
				users = make(map[int64]int)
			}

		//如果在10秒内没有100用户,则发一次
		case <-time.After(time.Duration(env.Config.Feed.PushGap) * time.Second):

			if len(users) > 0 {
				go SendNewFeedPush(users)
				users = make(map[int64]int)
			}
		}
	}

	return
}

func SendNewFeedPush(users map[int64]int) {

	for user, _ := range users {
		SendNewFeedMsg(user)
	}

	return
}
