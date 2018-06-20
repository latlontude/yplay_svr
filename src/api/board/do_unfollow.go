package board

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type UnfollowReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int64 `schema:"boardId"`
}

type UnfollowRsp struct {
	Code int `json:"code"` // 0 表示成功
}

func doUnfollow(req *UnfollowReq, r *http.Request) (rsp *UnfollowRsp, err error) {

	log.Debugf("uin %d, UnfollowReq %+v", req.Uin, req)

	code, err := Unfollow(req.Uin, req.BoardId)

	if err != nil {
		log.Errorf("uin %d, Unfollow error, %s", req.Uin, err.Error())
		return
	}

	rsp = &UnfollowRsp{code}

	log.Debugf("uin %d, UnfollowRsp succ, %+v", req.Uin, rsp)

	return
}

func Unfollow(uin, boardId int64) (code int, err error) {
	log.Debugf("start Unfollow uin:%d boardId:%d", uin, boardId)

	code = -1
	if uin <= 0 || boardId <= 0 {
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

	sql := fmt.Sprintf(`update v2follow  set status = 1 where uin = %d and boardId = %d`, uin, boardId)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	code = 0
	log.Debugf("end Unfollow uin:%d code:%d", uin, code)
	return
}
