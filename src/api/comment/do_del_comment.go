package comment

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelCommentReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId  int `schema:"answerId"`
	CommentId int `schema:"commentId"`
}

type DelCommentRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelComment(req *DelCommentReq, r *http.Request) (rsp *DelCommentRsp, err error) {

	log.Debugf("uin %d, DelCommentReq %+v", req.Uin, req)

	code, err := DelComment(req.Uin, req.AnswerId, req.CommentId)

	if err != nil {
		log.Errorf("uin %d, DelAnswer error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelCommentRsp{code}

	log.Debugf("uin %d, DelCommentRsp succ, %+v", req.Uin, rsp)

	return
}

func DelComment(uin int64, answerId, commentId int) (code int, err error) {
	log.Debugf("start DelComment uin = %d answerId = %d commentId = %d", uin, answerId, commentId)

	code = -1

	if answerId <= 0 || commentId <= 0 {
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

	sql := fmt.Sprintf(`delete from v2comments where ownerUid = %d and answerId = %d and commentId = %d`, uin, answerId, commentId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	code = 0

	log.Debugf("end DelComment uin = %d  code = %d", uin, code)
	return
}
