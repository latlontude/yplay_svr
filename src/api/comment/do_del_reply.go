package comment

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelReplyReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	AnswerId  int `schema:"answerId"`
	CommentId int `schema:"commentId"`
	ReplyId   int `schema:"replyId"`
}

type DelReplyRsp struct {
	Code int `json:"code"` // 0表示成功
}

func doDelReply(req *DelReplyReq, r *http.Request) (rsp *DelReplyRsp, err error) {

	log.Debugf("uin %d, DelReplyReq %+v", req.Uin, req)

	code, err := DelReply(req.Uin, req.AnswerId, req.CommentId, req.ReplyId)

	if err != nil {
		log.Errorf("uin %d, DelReply error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelReplyRsp{code}

	log.Debugf("uin %d, DelReplyRsp succ, %+v", req.Uin, rsp)

	return
}

func DelReply(uin int64, answerId, commentId, replyId int) (code int, err error) {
	log.Debugf("start DelReply uin = %d", uin)

	code = -1

	if answerId <= 0 || commentId <= 0 || replyId <= 0 {
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

	sql := fmt.Sprintf(`delete from v2replys where fromUid = %d and replyId = %d`, uin, replyId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	code = 0

	log.Debugf("end DelReplay uin = %d code = %d", uin, code)
	return
}
