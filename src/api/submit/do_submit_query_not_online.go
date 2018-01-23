package submit

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type SubmitQueryListNotOnlineReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type SubmitQueryListNotOnlineRsp struct {
	Infos []*SubmitQInfo `json:"infos"`
}

func doSubmitQueryListNotOnline(req *SubmitQueryListNotOnlineReq, r *http.Request) (rsp *SubmitQueryListNotOnlineRsp, err error) {

	log.Debugf("uin %d, SubmitQueryListNotOnlineReq %+v", req.Uin, req)

	infos, err := SubmitQueryListNotOnline(req.Uin)
	if err != nil {
		log.Errorf("uin %d, SubmitQueryListNotOnlineRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitQueryListNotOnlineRsp{infos}

	log.Debugf("uin %d, SubmitQueryListNotOnlineRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitQueryListNotOnline(uin int64) (infos []*SubmitQInfo, err error) {

	infos = make([]*SubmitQInfo, 0)

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

	sql := fmt.Sprintf(`select id, qid, qtext, qiconId, status, descr from submitQuestions where uin = %d and (status = %d or status = %d) order by ts desc`, uin, 0, 2)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info SubmitQInfo
		var qiconId int

		rows.Scan(&info.SubmitId, &info.QId, &info.QText, &qiconId, &info.Status, &info.Desc)
		info.QIconUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/qicon/%d.png", qiconId)

		infos = append(infos, &info)
	}

	return
}
