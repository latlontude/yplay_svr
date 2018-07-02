package like

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
)

type DelLikeReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid    int `schema:"qid"`
	LikeId int `schema:"likeId"`
	Typ    int `schema:"type"`
}

type DelLikeRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelLike(req *DelLikeReq, r *http.Request) (rsp *DelLikeRsp, err error) {

	log.Debugf("uin %d, DelLikeReq %+v", req.Uin, req)

	code, err := DelLike(req.Uin, req.Qid, req.LikeId, req.Typ)

	if err != nil {
		log.Errorf("uin %d, DelLike error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelLikeRsp{code}

	log.Debugf("uin %d, DelLikeRsp succ, %+v", req.Uin, rsp)

	return
}

func DelLike(uin int64, qid, likeId, typ int) (code int, err error) {
	log.Debugf("start DelLike uin = %d", uin)

	code = -1

	if uin <= 0 || qid <= 0 || likeId <= 0 || typ <= 0 || typ > 3 {
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


	err =updateV2Like(uin,qid,typ,likeId,2,inst)
	code = 0
	log.Debugf("end DelLike uin = %d ", uin)

	return
}
