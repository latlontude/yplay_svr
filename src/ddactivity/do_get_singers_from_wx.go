package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetSingerFromWxReq struct {
	OpenId string `schema:"openId"`
}

type GetSingerFromWxRsp struct {
	Singers       []SingerInfo `json:"singers"`
	VotedSingerId int          `json:"votedSingerId"`
}

func doGetSingersFromWx(req *GetSingerFromWxReq, r *http.Request) (rsp *GetSingerFromWxRsp, err error) {
	log.Debugf("start doGetSingersFromWx openId:%s", req.OpenId)
	singers, err := GetSingersFromWx(req.OpenId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	singerId, err := GetMyVotedSingerIdFromWx(req.OpenId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetSingerFromWxRsp{singers, singerId}
	log.Debugf("end doGetSingersFromWx rsp:%+v", rsp)
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

	sql := fmt.Sprintf(`select singerId, uin, activeHeadImgUrl, singerDetailInfoImgUrl, deptName, songName, songStoreName, songDuration  from ddsingers where status = 0`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

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

	log.Debugf("end GetSingersFromWx singers:%+v", singers)
	return
}

func GetMyVotedSingerIdFromWx(openId string) (singerId int, err error) {
	log.Debugf("start GetMyVotedSingerIdFromWx openId:%s", openId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select singerId from ddsingerFansFromWx where status = 0 and openId = "%s"`, openId)
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

	log.Debugf("end GetMyVotedSingerIdFromWx singerId:%d", singerId)
	return
}
