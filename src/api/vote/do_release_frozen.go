package vote

import (
	"api/im"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type ReleaseFrozenReq struct {
	User int64 `schema:"user"`
}

type ReleaseFrozenRsp struct {
}

func doReleaseFrozen(req *ReleaseFrozenReq, r *http.Request) (rsp *ReleaseFrozenRsp, err error) {

	log.Errorf("uin %d, ReleaseFrozenReq %+v", req.User, req)

	err = ReleaseFrozen(req.User)
	if err != nil {
		log.Errorf("uin %d, ReleaseFrozenRsp error, %s", req.User, err.Error())
		return
	}

	rsp = &ReleaseFrozenRsp{}

	log.Errorf("uin %d, ReleaseFrozenRsp succ, %+v", req.User, rsp)

	return
}

func ReleaseFrozen(uin int64) (err error) {
	log.Debugf("start ReleaseFrozen uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	now := int(time.Now().Unix())
	sql := fmt.Sprintf(`update freezingStatus set freezeTs = 0, ts = %d where uin = %d`, now, uin)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Error(err.Error())
		return
	}

	content := "开始新一轮投票吧(๑‾ ꇴ ‾๑)"
	err = im.SendLeaveFrozenMsg(uin, content)
	if err != nil {
		log.Error(err.Error())
	}

	log.Debugf("end ReleaseFrozen")
	return
}
