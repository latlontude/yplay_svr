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

type GetUserGemStatInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	UserUin  int64 `schema:"userUin"`
	PageNum  int   `schema:"pageNum"`
	PageSize int   `schema:"pageSize"`
}

type GetUserGemStatInfoRsp struct {
	Total int                   `json:"total"`
	Stats []*st.UserGemStatInfo `json:"stats"`
}

func doGetUserGemStatInfo(req *GetUserGemStatInfoReq, r *http.Request) (rsp *GetUserGemStatInfoRsp, err error) {

	log.Debugf("uin %d, GetUserGemStatInfoReq %+v", req.Uin, req)

	total, stats, err := GetUserGemStatInfo(req.UserUin, req.PageNum, req.PageSize)
	if err != nil {
		log.Errorf("uin %d, GetUserGemStatInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetUserGemStatInfoRsp{total, stats}

	log.Debugf("uin %d, GetUserGemStatInfoRsp succ, %+v", req.Uin, rsp)

	return
}

func GetUserGemStatInfo(uin int64, pageNum, pageSize int) (total int, stats []*st.UserGemStatInfo, err error) {

	stats = make([]*st.UserGemStatInfo, 0)

	if uin == 0 {
		return
	}

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select statValue from userStat where uin = %d and statField = %d`, uin, constant.ENUM_USER_STAT_GEM_CNT)

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

	sql = fmt.Sprintf(`select qid, count(id) as cnt, max(ts) as t from voteRecords where voteToUin = %d group by qid order by cnt desc, t desc limit %d, %d`, uin, s, e)

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
