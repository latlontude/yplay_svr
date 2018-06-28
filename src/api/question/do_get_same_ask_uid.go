package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"strings"
)



type GetSameAskRsp struct {
	Uids []string `json:"uids"`
	Code int      `json:"code"` // 0表示成功
}

func doGetSameAskUid(req *SameAskReq, r *http.Request) (rsp *GetSameAskRsp, err error) {

	log.Debugf("uin %d, SameAskReq %+v", req.Uin, req)

	uids ,code, err := GetSameAskUidArr(req.Qid)

	if err != nil {
		log.Errorf("uin %d, GetSameAskUid error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetSameAskRsp{uids,code}

	log.Debugf("uin %d, GetSameAskUid succ, %+v", req.Uin, rsp)

	return
}



func GetSameAskUidArr(qid int) (uids []string, code int ,err error) {
	if qid == 0 {
		log.Errorf("qid is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select sameAskUid from  v2questions where qid = %d`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var sameAskUid string
	for rows.Next() {
		rows.Scan(&sameAskUid)
	}
	uidArr := strings.Split(sameAskUid,",")
	for _, uid := range uidArr {
		if uid == "" {
			continue
		}
		uids = append(uids, uid)
	}

	return
}