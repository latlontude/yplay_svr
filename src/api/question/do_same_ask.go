



//同问 更新sameAskUid字段的值


package question


import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type SameAskReq struct {
	Uin   		int64  `schema:"uin"`
	Token 		string `schema:"token"`
	Ver   		int    `schema:"ver"`
	Qid         int    `schema:"qid"`

}

type SameAskRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doSameAsk(req *SameAskReq, r *http.Request) (rsp *SameAskRsp, err error) {

	log.Debugf("uin %d, SameAskReq %+v", req.Uin, req)

	code, err := SameAskQuestion(req.Uin, req.Qid)

	if err != nil {
		log.Errorf("uin %d, SameAskReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SameAskRsp{code}

	log.Debugf("uin %d, SameAskReq succ, %+v", req.Uin, rsp)

	return
}

func SameAskQuestion(uin int64, qid int) (code int, err error) {
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
	//追加同问uid
	sql := fmt.Sprintf(`update v2questions set sameAskUid =  CONCAT(sameAskUid,',%d')  where qid = %d`, uin, qid)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}
	code = 0
	return
}



