package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type ModifyQuestionReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid         int    `schema:"qid"`
	QTitle      string `schema:"qTitle"`
	QContent    string `schema:"qContent"`
	QImgUrls    string `schema:"qImgUrls"`
	IsAnonymous int    `schema:"isAnonymous"`
}

type ModifyQuestionRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doModifyQuestion(req *ModifyQuestionReq, r *http.Request) (rsp *ModifyQuestionRsp, err error) {

	log.Debugf("uin %d, ModifyQuestionReq %+v", req.Uin, req)

	code, err := ModifyQuestion(req.Uin, req.Qid, req.QTitle, req.QContent, req.QImgUrls, req.IsAnonymous)

	if err != nil {
		log.Errorf("uin %d, ModifyQuestionReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ModifyQuestionRsp{code}

	log.Debugf("uin %d, ModifyQuestionRsp succ, %+v", req.Uin, rsp)

	return
}

func ModifyQuestion(uin int64, qid int, qTitle, qContent, qImgUrls string, isAnonymous int) (code int, err error) {
	log.Debugf("start ModifyQuestion uin = %d qid = %d", uin, qid)

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

	ts := time.Now().Unix()
	sql := fmt.Sprintf(`update v2questions set qTitle = '%s',
                                           qContent = '%s',
                                           qImgUrls = '%s',
                                           isAnonymous = %d,
                                           modTs = %d
                                           where ownerUid = %d and qid = %d`,
		qTitle, qContent, qImgUrls, isAnonymous, ts, uin, qid)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	code = 0

	log.Debugf("end ModifyQuestion uin = %d qid = %d code = %d", uin, qid, code)
	return
}
