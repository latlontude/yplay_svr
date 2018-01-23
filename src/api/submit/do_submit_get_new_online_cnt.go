package submit

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
)

type SubmitGetNewOnlineCntReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type SubmitGetNewOnlineCntRsp struct {
	Cnt int `json:"cnt"`
}

func doSubmitGetNewOnlineCnt(req *SubmitGetNewOnlineCntReq, r *http.Request) (rsp *SubmitGetNewOnlineCntRsp, err error) {

	log.Errorf("uin %d, SubmitGetNewOnlineCntReq %+v", req.Uin, req)

	cnt, err := SubmitGetNewOnlineCnt(req.Uin)
	if err != nil {
		log.Errorf("uin %d, SubmitGetNewOnlineCntRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitGetNewOnlineCntRsp{cnt}

	log.Errorf("uin %d, SubmitGetNewOnlineCntRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitGetNewOnlineCnt(uin int64) (cnt int, err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SUBMIT_LAST_READ_ONLINE_TS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//已经上线的，判断是否新上线的问题列表
	//对比上次的拉取时间
	keyStr := fmt.Sprintf("%d", uin)
	valStr, err := app.Get(keyStr)

	if err != nil {

		//如果KEY不存在 则认为lastMsgId为0
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				valStr = "0"
			} else {
				log.Errorf(err.Error())
				return
			}
		} else {
			log.Errorf(err.Error())
			return
		}
	}

	lastTs, err1 := strconv.Atoi(valStr)
	if err1 != nil {
		log.Errorf(err1.Error())
		lastTs = 0
	}

	sql := fmt.Sprintf(`select count(id) from submitQuestions where uin = %d and status = %d and mts > %d`, uin, 1, lastTs)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&cnt)
	}

	return
}
