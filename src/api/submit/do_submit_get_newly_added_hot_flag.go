package submit

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
	"svr/st"
)

type SubmitGetNewlyAddedHotFlagReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type SubmitGetNewlyAddedHotFlagRsp struct {
	Flag int `json:"flag"`
}

func doSubmitGetNewlyAddedHotFlag(req *SubmitGetNewlyAddedHotFlagReq, r *http.Request) (rsp *SubmitGetNewlyAddedHotFlagRsp, err error) {

	log.Errorf("uin %d, SubmitGetNewlyAddedHotFlagReq %+v", req.Uin, req)

	flag, err := SubmitGetNewlyAddedHotFlag(req.Uin)
	if err != nil {
		log.Errorf("uin %d, SubmitGetNewlyAddedHotFlagRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitGetNewlyAddedHotFlagRsp{flag}

	log.Errorf("uin %d, SubmitGetNewlyAddedHotFlagRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitGetNewlyAddedHotFlag(uin int64) (flag int, err error) {

	log.Errorf("start submitGetNewlyAddedHotFlag")

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

	//获取用户投稿上线题目列表
	sql := fmt.Sprintf(`select qid from submitQuestions where uin = %d and status = 1 order by mts desc`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	qids := make([]int, 0)

	for rows.Next() {
		var qid int
		rows.Scan(&qid)
		qids = append(qids, qid)
	}

	if len(qids) == 0 {
		return
	}

	qidsStr := ""
	for _, qid := range qids {
		qidsStr += fmt.Sprintf("%d,", qid)
	}
	qidsStr = qidsStr[:len(qidsStr)-1]

	//获取上次拉取时间
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

	sql = fmt.Sprintf(`select voteToUin from voteRecords where qid in (%s) and ts > %d group by voteToUin`, qidsStr, lastTs)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	uinsSlice := make([]int64, 0)
	in := false
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		uinsSlice = append(uinsSlice, uid)
		if uid == uin {
			in = true
		}
	}

	if len(uinsSlice) == 0 {
		flag = 0
		return
	}

	if !in {
		uinsSlice = append(uinsSlice, uin)
	}

	res, err := st.BatchGetUserProfileInfo(uinsSlice)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uid := range uinsSlice {
		if _, ok := res[uid]; ok {
			if res[uid].SchoolId == res[uin].SchoolId && res[uid].Grade == res[uin].Grade {

				if uid == uin {
					if in {
						flag = 1
						break
					}
				} else {
					flag = 1
					break
				}
			}
		}
	}
	log.Errorf("end submitGetNewlyAddedHotFlag")
	return
}
