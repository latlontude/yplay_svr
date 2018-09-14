package question

import (
	_ "api/common"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	_ "sort"
	_ "svr/st"
)

type GetPoiTagListReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int `schema:"boardId"`

	Version int `schema:"version"`
}

type PoiInfo struct {
	Tag       string  `schema:"tag"`
	Longitude float64 `schema:"longitude"`
	Latitude  float64 `schema:"latitude"`
}

type GetPoiTagListRsp struct {
	Tags []*PoiInfo `json:"poitags"`
}

func doGetPoiTagList(req *GetPoiTagListReq, r *http.Request) (rsp *GetPoiTagListRsp, err error) {

	log.Debugf("uin %d, GetPoiTagListReq %+v", req.Uin, req)

	tags, err := GetPoiTagList(req.Uin, req.BoardId, 0, 100, req.Version)

	if err != nil {
		log.Errorf("uin %d, GetPoiTagList error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetPoiTagListRsp{tags}

	log.Debugf("uin %d, GetPoiTagListRsp succ  , rsp:%v", req.Uin, rsp)

	return
}

func GetPoiTagList(uin int64, boardId, pageNum, pageSize, version int) (tags []*PoiInfo, err error) {

	//log.Debugf("start GetQuestions uin:%d", uin)

	if boardId <= 0 || pageNum < 0 || pageSize < 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}
	tags = make([]*PoiInfo, 0)

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	offset := (pageNum - 1) * pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(1) num,avg(longitude) lng,avg(latitude) lat,poiTag from v2questions 
						where boardId=%d and poiTag != '' group by poiTag order by modTs desc limit %d,%d`,
		boardId, offset, pageSize)

	log.Errorf("SQL:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	tags = make([]*PoiInfo, 0)
	for rows.Next() {
		var num int64
		var poi PoiInfo

		rows.Scan(&num, &poi.Longitude, &poi.Latitude, &poi.Tag)
		tags = append(tags, &poi)
	}

	//log.Debugf("end GetQuestions uin:%d totalCnt:%d", uin, totalCnt)
	return
}
