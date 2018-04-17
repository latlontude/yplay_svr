package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type SingerInfo struct {
	SingerId               int    `json:"singerId"`
	Uin                    int64  `json:"uin"`
	UserName               string `json:"userName"`
	NickName               string `json:"nickName"`
	HeadImgUrl             string `json:"headImgUrl"`
	Gender                 int    `json:"gender"`
	DeptName               string `json:"deptName"`
	Grade                  int    `json:"grade"`
	ActiveHeadImgUrl       string `json:"activeHeadImgUrl"`
	SingerDetailInfoImgUrl string `json:"singerDetailInfoImgUrl"`
	SingerSongUrl          string `json:"songUrl"`
	SingerSongName         string `json:"songName"`
	SingerSongDuration     string `json:"songDuration"`
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

	sql := fmt.Sprintf(`select singerId, uin, activeHeadImgUrl, singerDetailInfoImgUrl, deptName, songName, songStoreName, songDuration  from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singer SingerInfo
		var songStoreName string
		rows.Scan(&singer.SingerId, &singer.Uin, &singer.ActiveHeadImgUrl, &singer.SingerDetailInfoImgUrl, &singer.DeptName, &singer.SingerSongName, &songStoreName, &singer.SingerSongDuration)
		singer.ActiveHeadImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singer.ActiveHeadImgUrl)
		singer.SingerDetailInfoImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singer.SingerDetailInfoImgUrl)
		singer.SingerSongUrl = "https://ddactive.yeejay.com/music/" + songStoreName

		ui, err1 := st.GetUserProfileInfo(singer.Uin)
		if err1 != nil {
			err = err1
			log.Errorf(err1.Error())
			return
		}

		singer.UserName = ui.UserName
		singer.NickName = ui.NickName
		singer.HeadImgUrl = ui.HeadImgUrl
		singer.Gender = ui.Gender
		singer.Grade = ui.Grade

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

func uinExist(uin int64) (pass bool, err error) {
	log.Debugf("start uinExist uin:%d", uin)

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

	//看看是否在配置的学校列表中
	log.Debugf("openSchools:%+v", OpenSchools)
	log.Debugf("ui schoolId:%d", ui.SchoolId)

	find := false
	for sid, _ := range OpenSchools {
		if ui.SchoolId == sid {
			find = true
			break
		}
	}

	//没有参加活动的权限
	if !find {
		pass = false
		log.Debugf("user:%d school does not fit", uin)
		return
	}

	pass = true
	log.Debugf("end uinExist pass : %+t", pass)
	return
}
