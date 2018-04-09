package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type SingerInfo struct {
	SingerId   int    `json:"singerId"`
	Uin        int    `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	DeptName   string `json:"deptName"`
	Grade      int    `json:"grade"`
}

type GetSingerFromPupuReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetSingerFromPupuRsp struct {
	Singers       []SingerInfo `json:"singers"`
	VotedSingerId int          `json:"votedSingerId"`
}

func doGetSingersFromPupu(req *GetSingerFromPupuReq, r *http.Request) (rsp *GetSingerFromPupuRsp, err error) {
	log.Debugf("start doGetSingersFromPupu uin:%d", req.Uin)
	singers, err := GetSingersFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	singerId, err := GetMyVotedSingerIdFromPupu(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerFromPupuRsp{singers, singerId}
	log.Debugf("end doGetSingersFromPupu rsp:%+v", rsp)
	return
}

func GetSingersFromPupu(uin int64) (singers []SingerInfo, err error) {
	log.Debugf("start GetSingersFromPupu uin:%d", uin)

	singers = make([]SingerInfo, 0)

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
		var singer SingerInfo
		rows.Scan(&singer.SingerId, &singer.Uin, &singer.NickName, &singer.HeadImgUrl, &singer.Gender, &singer.DeptName, &singer.Grade)
		singers = append(singers, singer)
	}

	log.Debugf("end GetSingersFromPupu singers:%+v", singers)
	return
}

func GetMyVotedSingerIdFromPupu(uin int64) (singerId int, err error) {
	log.Debugf("start GetMyVotedSingerIdFromPupu uin:%d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId from ddsingerFansFromPupu where status = 0 and uin = %d`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&singerId)
	}

	log.Debugf("end GetMyVotedSingerIdFromPupu singerId:%d", singerId)
	return
}
