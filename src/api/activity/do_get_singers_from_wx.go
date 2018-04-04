package activity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type GetSingerFromWxReq struct {
	OpenId string `schema:"openId"`
}

type GetSingerFromWxRsp struct {
	Singers []SingerInfo `json:"singers"`
}

func doGetSingerFromWx(req *GetSingerFromWxReq, r *http.Request) (rsp *GetSingerFromWxRsp, err error) {
	log.Debugf("start doGetSingerFromWx openId:%s", req.OpenId)
	singers, err := GetSingersFromWx(req.OpenId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerFromWxRsp{singers}
	log.Debugf("end doGetSingerFromWx")
	return
}

func GetSingersFromWx(openId string) (singers []SingerInfo, err error) {
	log.Debugf("start GetSingersFromWx openId:%s", openId)

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

	log.Debugf("end GetSingersFromWx singers:%+v", singers)
	return
}
