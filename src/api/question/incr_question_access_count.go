package question

import (
	"common/constant"
	"common/myredis"
	"fmt"
)

func buildKey(qid int) (key string) {
	key = fmt.Sprintf(`question_%d`, qid)
	return
}

//增加一个问题的访问数量
func IncreaseQuestionAccessCount(qid int) (count int, err error) {
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_2DEGREE_FRIENDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	key := buildKey(qid)
	count, err = app.Incr(key)
	if err != nil {
		log.Error(err.Error())
		return
	}
	return
}
