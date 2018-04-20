package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/cache"
	"svr/st"
)

type GetUserGemStatInfoByCntReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Hide    int   `schema:"hide"` // 0 不隐藏，1 隐藏
	UserUin int64 `schema:"userUin"`
	Start   int   `schema:"start"` //这个数量之后的
	Cnt     int   `schema:"cnt"`   // 获取多少条
}

type GetUserGemStatInfoByCntRsp struct {
	Total int                   `json:"total"`
	Stats []*st.UserGemStatInfo `json:"stats"`
}

func doGetUserGemStatInfoByCnt(req *GetUserGemStatInfoByCntReq, r *http.Request) (rsp *GetUserGemStatInfoByCntRsp, err error) {

	log.Debugf("uin %d, doGetUserGemStatInfoByCnt %+v", req.Uin, req)

	if req.Uin != req.UserUin && req.Hide != 0 {
		err = rest.NewAPIError(constant.E_DB_QUERY, "invalid params")
		log.Errorf("query other people diamond infomation,but hide field is not zero")
		return
	}

	total, stats, err := GetUserGemStatInfoByCnt(req.UserUin, req.Hide, req.Start, req.Cnt)
	if err != nil {
		log.Errorf("uin %d, GetUserGemStatInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetUserGemStatInfoByCntRsp{total, stats}

	log.Debugf("uin %d, doGetUserGemStatInfoByCnt succ, %+v", req.Uin, rsp)

	return
}

func GetUserGemStatInfoByCnt(uin int64, hide, start, cnt int) (total int, stats []*st.UserGemStatInfo, err error) {

	stats = make([]*st.UserGemStatInfo, 0)

	if uin == 0 {
		return
	}

	if start < 0 || cnt > 10 || cnt < 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select count(id) as cnt from voteRecords where voteToUin = %d and hide = %d`, uin, hide)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	sql = fmt.Sprintf(`select qid, count(id) as cnt, max(ts) as t from voteRecords where voteToUin = %d and hide = %d group by qid order by cnt desc, t desc limit %d, %d`, uin, hide, start, cnt)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var info st.UserGemStatInfo
		var tmp int
		rows.Scan(&info.QId, &info.GemCnt, &tmp)

		info.Uin = uin

		if q, ok := cache.QUESTIONS[info.QId]; ok {
			info.QText = q.QText
			info.QIconUrl = q.QIconUrl
		}

		stats = append(stats, &info)
	}

	return
}

func GetUserGemCnt(uin int64) (total int, err error) {
	log.Debugf("start GetUserGemCnt uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select count(id) as cnt from voteRecords where voteToUin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	log.Debugf("end GetUserGemCnt total:%d", total)
	return
}
