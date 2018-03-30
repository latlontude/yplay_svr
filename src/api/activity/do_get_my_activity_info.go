package activity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetMyActivityInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type BannerInfo struct {
	ImgUrl    string            `json:"imgUrl"`
	ImgWidth  int               `json:"imgWidth"`
	ImgHeight int               `json:"imgHeight"`
	H5Url     string            `json:"h5Url"`
	Params    map[string]string `json:"params"`
	TsStart   int               `json:"tsStart"`
	TsEnd     int               `json:"tsEnd"`
}

type GetMyActivityInfoRsp struct {
	BannerOpen int           `json:"bannerOpen"`
	Banners    []*BannerInfo `json:"banners"`
}

func doGetMyActivityInfo(req *GetMyActivityInfoReq, r *http.Request) (rsp *GetMyActivityInfoRsp, err error) {

	log.Errorf("uin %d, GetMyActivityInfoReq %+v", req.Uin, req)

	open, banners, err := GetMyActivityInfo(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetMyActivityInfoRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetMyActivityInfoRsp{open, banners}

	log.Errorf("uin %d, GetMyActivityInfoRsp succ, %+v", req.Uin, rsp)

	return
}

func GetMyActivityInfo(uin int64) (open int, banners []*BannerInfo, err error) {

	open = 0
	banners = make([]*BannerInfo, 0)

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//没有参加活动的权限
	if ui.SchoolId != 1 {
		return
	}

	sql := fmt.Sprintf(`select imgUrl, imgWidth, imgHeight, h5Url, tsStart, tsEnd from banner`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	params := make(map[string]string)
	params["a"] = "100"
	params["b"] = "200"

	for rows.Next() {

		var bi BannerInfo
		bi.Params = params

		rows.Scan(&bi.ImgUrl, &bi.ImgWidth, &bi.ImgHeight, &bi.H5Url, &bi.TsStart, &bi.TsEnd)
		banners = append(banners, &bi)
	}

	if len(banners) > 0 {
		open = 1
	}

	return
}
