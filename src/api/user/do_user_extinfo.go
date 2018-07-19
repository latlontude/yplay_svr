package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type ExtUserInfoReq struct {
	Uin   int64  `schema:"uin"`
	FUin  int64  `schema:"friendUin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type ExtUserInfoRsp struct {
	QuestionCnt int `json:"questionCnt"`
	AnswerCnt   int `json:"answerCnt"`
}

func doExUserInfo(req *ExtUserInfoReq, r *http.Request) (rsp *ExtUserInfoRsp, err error) {

	log.Errorf("uin %d, ExtUserInfoReq %+v", req.Uin, req)

	queryUin := req.Uin

	//查询好友extra
	if req.FUin > 0 {
		queryUin = req.FUin
	}

	qstCnt, answerCnt, err := ExtInfo(queryUin)
	if err != nil {
		log.Errorf("uin %d, ExtUserInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &ExtUserInfoRsp{qstCnt, answerCnt}

	log.Errorf("uin %d, ExtUserInfoRsp succ, %+v", req.Uin, rsp)

	return
}

func ExtInfo(uin int64) (qstCnt int, answerCnt int, err error) {

	if uin == 0 {
		log.Error("uin is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(*) from  v2answers ,v2questions 
			where v2answers.answerStatus = 0 
			and v2questions.qStatus = 0 
			and v2answers.qid=v2questions.qid 
			and v2answers.ownerUid = %d`, uin)

	rows, err := inst.Query(sql)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&answerCnt)
	}

	sql = fmt.Sprintf(`select count(*)  from v2questions where ownerUid=%d and qStatus = 0 `, uin)
	rows, err = inst.Query(sql)
	for rows.Next() {
		rows.Scan(&qstCnt)
	}

	return

}
