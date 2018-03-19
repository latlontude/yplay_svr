package im

import (
	"common/env"
	"time"
)

func NewStoryPushRoutine() {

	users := make(map[int64]int)

	for {
		select {
		case uin := <-ChanStoryPush:
			users[uin] = 1

		case <-time.After(time.Duration(env.Config.Story.PushGap) * time.Second):

			if len(users) > 0 {
				go SendNewStoryPush(users)
				users = make(map[int64]int)
			}
		}
	}

	return
}

func SendNewStoryPush(users map[int64]int) {

	for user, _ := range users {
		SendNewStoryMsg(user)
	}

	return
}
