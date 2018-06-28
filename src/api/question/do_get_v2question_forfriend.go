package question

import (
	"net/http"
)



type GetV2QuestionsForFriendReq struct {
	Uin      int64  `schema:"uin"`
	FUin     int64  `schema:"fuin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}



func doGetV2QuestionsForFriend(req *GetV2QuestionsForFriendReq, r *http.Request) (rsp *GetV2QuestionsRsp, err error) {

	log.Debugf("uin %d,fuid %d , GetQuestionsReq %+v", req.Uin,req.FUin, req)

	//我提出的问题
	questions, totalCnt, qstCnt,answerCnt,err := GetV2QuestionsAndAnswer(req.FUin, req.PageSize, req.PageNum)

	if err != nil {
		log.Errorf("uin %d, doGetV2QuestionsForFriend error, %s", req.FUin, err.Error())
		return
	}

	rsp = &GetV2QuestionsRsp{questions, totalCnt,qstCnt,answerCnt}

	log.Debugf("uin %d fuin:%d, doGetV2QuestionsForFriend success", req.Uin,req.FUin)

	return
}

