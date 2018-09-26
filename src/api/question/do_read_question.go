package question

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
)

type ReadQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid     int `schema:"qid"`
	Version int `schema:"version"`
}

type ReadQuestionRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doReadQuestion(req *ReadQuestionReq, r *http.Request) (rsp *ReadQuestionRsp, err error) {

	log.Debugf("uin %d, ReadQuestionReq %+v", req.Uin, req)

	code, err := ReadQuestion(req.Uin, req.Qid, req.Version)

	if err != nil {
		log.Errorf("uin %d, ReadQuestion error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ReadQuestionRsp{code}

	log.Debugf("uin %d, ReadQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func ReadQuestion(uin int64, qid int, version int) (code int, err error) {

	if uin <= 0 || qid <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
	if err != nil {
		log.Errorf(err.Error())
		//return
	}
	key := fmt.Sprintf("read_question_list:%d", uin)
	value, err := app.Get(key)

	log.Debugf("key:%s value:%s", key, value)
	return
}
