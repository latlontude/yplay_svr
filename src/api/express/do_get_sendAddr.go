package express

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type GetAddrListReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SchoolId int    `schema:"schoolId"`
}
type AddrInfo struct {
	Area      string `json:"area"`
	PointName string `json:"pointName"`
}
type GetAddrListRsp struct {
	AddrList []*AddrInfo `json:"addrList"`
}

func doGetAddrList(req *GetSenderListReq, r *http.Request) (rsp *GetAddrListRsp, err error) {

	log.Debugf("GetBoardSenderListReq req:%+v", req)
	addrList, err := GetAddrList(req.Uin, req.SchoolId)

	if err != nil {
		log.Debugf("GetBoardSenderList error ,err:%+v", err)
		return
	}

	rsp = &GetAddrListRsp{addrList}

	log.Debugf("GetBoardSenderListRsp , rsp:%+v", rsp)

	return
}

func GetAddrList(uin int64, schoolId int) (addrList []*AddrInfo, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select area,pointName from express_points where schoolId = %d `, schoolId)

	log.Debugf("express_points sql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	addrList = make([]*AddrInfo, 0)
	for rows.Next() {
		var addrInfo AddrInfo
		rows.Scan(&addrInfo.Area, &addrInfo.PointName)
		addrList = append(addrList, &addrInfo)
	}
	return
}
