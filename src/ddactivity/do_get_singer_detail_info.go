package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type SingerDetailInfo struct {
	SingerId   int    `json:"singerId"`
	Uin        int    `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	DeptName   string `json:"deptName"`
	Grade      int    `json:"grade"`
}

type GetSingerDetailInfoFromPupuReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SingerId int    `schema:"singerId"`
}

type GetSingerDetailInfoFromPupuRsp struct {
	SingerWithDetailInfo SingerDetailInfo `json:"singerWithDetailInfo"`
}

func doGetSingerDetailInfoFromPupu(req *GetSingerDetailInfoFromPupuReq, r *http.Request) (rsp *GetSingerDetailInfoFromPupuRsp, err error) {
	log.Debugf("start doGetSingerDetailInfoFromPupu uin:%d", req.Uin)
	singerWithDetailInfo, err := GetSingerDetailInfoFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerDetailInfoFromPupuRsp{singerWithDetailInfo}
	log.Debugf("end doGetSingerDetailInfoFromPupu rsp:%+v", rsp)
	return
}

func GetSingerDetailInfoFromPupu(uin int64) (singerWithDetailInfo SingerDetailInfo, err error) {
	log.Debugf("start GetSingerDetailInfoFromPupu uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId, uin, nickName, headImgUrl, gender, deptName, grade from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&singerWithDetailInfo.SingerId, &singerWithDetailInfo)
	}

	log.Debugf("end GetSingerDetailInfoFromPupu singerWithDetailInfo:%+v", singerWithDetailInfo)
	return
}
