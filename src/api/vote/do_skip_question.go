package vote

import (
	"net/http"
	"svr/st"
)

type SkipQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	QId   int `schema:"qid"`
	Index int `schema:"index"`
}

type SkipQuestionRsp struct {
}

func doSkipQuestion(req *SkipQuestionReq, r *http.Request) (rsp *SkipQuestionRsp, err error) {

	log.Debugf("uin %d, SkipQuestionReq %+v", req.Uin, req)

	//if req.Uin == 100328 || req.Uin == 100446 {
	if true {
		err = SkipQuestionPreGene(req.Uin, req.QId, req.Index)
	} else {
		err = SkipQuestion(req.Uin, req.QId, req.Index)
	}

	if err != nil {
		log.Errorf("uin %d, SkipQuestionRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SkipQuestionRsp{}

	log.Debugf("uin %d, SkipQuestionRsp succ, %+v", req.Uin, rsp)

	go UserActRecords(req.Uin, req.QId, 0)

	return
}

func SkipQuestion(uin int64, qid int, index int) (err error) {

	err = st.UpdateVoteProgress2(uin, qid, index)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}

func SkipQuestionPreGene(uin int64, qid int, index int) (err error) {

	err = st.UpdateVoteProgressByPreGene(uin, qid, index)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
