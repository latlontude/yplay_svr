package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type UpdateUserGemStatInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid  int `schema:"qid"`  //要更改隐藏状态的题目id
	Hide int `schema:"hide"` // 0 不隐藏，1 隐藏
}

type UpdateUserGemStatInfoRsp struct {
}

func doUpdateUserGemStatInfo(req *UpdateUserGemStatInfoReq, r *http.Request) (rsp *UpdateUserGemStatInfoRsp, err error) {

	log.Debugf("uin %d, doUpdateUserGemStatInfoReq %+v", req.Uin, req)

	err = UpdateUserGemStatInfo(req.Uin, req.Qid, req.Hide)
	if err != nil {
		log.Errorf("uin %d, UpdateUserGemStatInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &UpdateUserGemStatInfoRsp{}

	log.Debugf("uin %d, doUpdateUserGemStatInfoRsp succ, %+v", req.Uin, rsp)

	return
}

func UpdateUserGemStatInfo(uin int64, qid, hide int) (err error) {

	log.Debugf("start UpdateUserGemStatInfo uin:%d qid:%d hide:%d", uin, qid, hide)

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`update voteRecords set hide = %d where voteToUin = %d and qid = %d`, hide, uin, qid)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end UpdateUserGemStatInfo")
	return
}
