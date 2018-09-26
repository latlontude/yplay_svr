package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type GetHeatMapReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int `schema:"boardId"`
}

type Point struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
type GetHeatMapRsp struct {
	Points []*Point `json:"points"`
}

func doGetHeatMap(req *GetHeatMapReq, r *http.Request) (rsp *GetHeatMapRsp, err error) {
	log.Debugf("GetHeatMapReq:%+v", req)
	points, err := GetHeatMap(req.Uin, req.BoardId)

	if err != nil {
		return
	}

	rsp = &GetHeatMapRsp{points}
	log.Debugf("headMapRsp:%+v", rsp)

	return
}

func GetHeatMap(uin int64, boardId int) (points []*Point, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select longitude,latitude from  v2questions where boardId = %d and (longitude != 0 and latitude != 0) limit 1000`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	points = make([]*Point, 0)
	for rows.Next() {
		var point Point
		rows.Scan(&point.Longitude, &point.Latitude)
		points = append(points, &point)
	}
	return
}
