package answer

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelAnswerReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid      int `schema:"qid"`
	AnswerId int `schema:"answerId"`
}

type DelAnswerRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelAnswer(req *DelAnswerReq, r *http.Request) (rsp *DelAnswerRsp, err error) {

	log.Debugf("uin %d, DelAnswerReq %+v", req.Uin, req)

	code, err := DelAnswer(req.Uin, req.Qid, req.AnswerId)

	if err != nil {
		log.Errorf("uin %d, DelAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelAnswerRsp{code}

	log.Debugf("uin %d, DelAnswerRsp succ, %+v", req.Uin, rsp)

	return
}

func DelAnswer(uin int64, qid, answerId int) (code int, err error) {
	log.Debugf("start DelAnswer uin = %d qid = %d answerId = %d", uin, qid, answerId)

	code = -1

	if uin <= 0 || qid <= 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`delete from v2answers where ownerUid = %d and qid = %d and answerId = %d`, uin, qid, answerId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	code = 0

	log.Debugf("end DelAnswer uin = %d qid = %d answerId = %d code = %d", uin, qid, answerId, code)
	return
}
