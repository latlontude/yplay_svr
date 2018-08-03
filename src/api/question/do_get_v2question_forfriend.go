package question

import (
	"net/http"
)

type GetV2QuestionsForFriendReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	FUin     int64  `schema:"fuin"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

func doGetV2QuestionsForFriend(req *GetV2QuestionsForFriendReq, r *http.Request) (rsp *GetV2QuestionsRsp, err error) {

	log.Debugf("uin %d,fuid %d , GetQuestionsReq %+v", req.Uin, req.FUin, req)

	//我提出的问题
	questions, totalCnt, qstCnt, answerCnt, err := GetV2QuestionsAndAnswer(req.Uin, req.FUin, req.PageSize, req.PageNum)

	if err != nil {
		log.Errorf("uin %d, doGetV2QuestionsForFriend error, %s", req.FUin, err.Error())
		return
	}

	labelList, err := GetHomeLabelInfo(req.Uin, req.FUin)

	if err != nil {
		log.Errorf("uin %d, doGetV2QuestionsForMe error, %s", req.Uin, err.Error())
		return
	}

	loginDays, err := GetRegisterTime(req.Uin)

	if err != nil {
		log.Errorf("get profile register ts error", req.Uin, err.Error())
	}

	rsp = &GetV2QuestionsRsp{questions, totalCnt, qstCnt, answerCnt, labelList, loginDays}

	log.Debugf("uin %d fuin:%d, doGetV2QuestionsForFriend success", req.Uin, req.FUin)

	return
}
