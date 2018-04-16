package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type GetCallInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type CallInfo struct {
	Money                  int          `json:"money"`
	NeedCallCntToBeatFirst int          `json:"needCallCntToBeatFirst"`
	FirstSingers           []SingerInfo `json:"firstSingers"`
	MySinger               SingerInfo   `json:"mySinger"`
	LeftCallCnt            int          `json:"leftCallCnt"`
	FinishedTaskCnt        int          `json:"finishedTaskCnt"`
	TotalTaskCnt           int          `json:"totalTaskCnt"`
}

type GetCallInfoRsp struct {
	Status int      `json:"status"` // 0 其他错误，1 成功， 2 用户还未投票
	Info   CallInfo `json:"callInfo"`
}

func doGetCallInfo(req *GetCallInfoReq, r *http.Request) (rsp *GetCallInfoRsp, err error) {
	log.Debugf("start doGetCallInfo uin:%d", req.Uin)
	_, err = updateCallTypeInfos(req.Uin, []int{5, 6, 7})
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	ret, err := GetCallInfo(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &ret
	log.Debugf("end doGetCallInfo rsp:%+v", rsp)
	return
}

func GetCallInfo(uin int64) (ret GetCallInfoRsp, err error) {
	log.Debugf("start GetCallInfo uin:%d", uin)

	ret.Info.Money = Config.BonusPool.Money

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	// 获取得到打call次数最多的歌手及其得到的打call数
	sql := fmt.Sprintf(`select singerId, count(id) as cnt from ddcallForSingers where type = 8 group by singerId order by cnt desc`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	firstSingerIds := make([]int, 0)
	firstSingerIdCnt := 0
	for rows.Next() {
		var singerId int
		var cnt int
		rows.Scan(&singerId, &cnt)
		if cnt < firstSingerIdCnt {
			break
		}

		firstSingerIds = append(firstSingerIds, singerId)
		firstSingerIdCnt = cnt
	}

	// 查找我的爱豆
	sql = fmt.Sprintf(`select singerId from ddsingerFansFromPupu where status = 0 and uin = %d`, uin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	mySingerId := 0
	for rows.Next() {
		rows.Scan(&mySingerId)
	}

	if mySingerId == 0 {
		ret.Status = 2
		log.Debugf("uin:%d has not be any singer`s fans")
		return
	}

	singerIdsStr := ""
	for _, singerId := range firstSingerIds {
		singerIdsStr += fmt.Sprintf("%d,", singerId)
	}
	singerIdsStr += fmt.Sprintf("%d", mySingerId)

	//查找我的爱豆和获取得到打call次数最多的歌手的信息
	sql = fmt.Sprintf(`select singerId, uin,  activeHeadImgUrl, singerDetailInfoImgUrl, deptName  from ddsingers where singerId in (%s) and  status = 0`, singerIdsStr)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var singer SingerInfo
		rows.Scan(&singer.SingerId, &singer.Uin, &singer.ActiveHeadImgUrl, &singer.SingerDetailInfoImgUrl, &singer.DeptName)

		singer.ActiveHeadImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singer.ActiveHeadImgUrl)
		singer.SingerDetailInfoImgUrl = fmt.Sprintf("http://yplay-1253229355.cossh.myqcloud.com/banner/%s", singer.SingerDetailInfoImgUrl)

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

		if singer.SingerId == mySingerId {
			ret.Info.MySinger = singer
		}

		for _, singerId := range firstSingerIds {
			if singerId == singer.SingerId {
				ret.Info.FirstSingers = append(ret.Info.FirstSingers, singer)
				break
			}
		}

	}

	//获取我的爱豆得到的打call数
	sql = fmt.Sprintf(`select count(id) as cnt from ddcallForSingers where singerId = %d and type = 8 and status = 0`, mySingerId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	mySingerIdCnt := 0
	for rows.Next() {
		rows.Scan(&mySingerIdCnt)
	}

	find := false
	for _, singerId := range firstSingerIds {
		if singerId == mySingerId {
			find = true
			break
		}
	}

	if find {
		ret.Info.NeedCallCntToBeatFirst = firstSingerIdCnt - mySingerIdCnt
	} else {
		ret.Info.NeedCallCntToBeatFirst = firstSingerIdCnt - mySingerIdCnt + 1
	}

	//获取今日打call剩余次数
	t := time.Now()
	h, m, s := t.Clock()
	ts := t.Unix()

	minTs := ts - int64(h*3600+m*60+s)
	maxTs := minTs + int64(24*3600)

	sql = fmt.Sprintf(`select count(id) as cnt from ddcallForSingers where uin = %d and type = 8 and ts >= %d and ts <= %d`, uin, minTs, maxTs) //type = 8 为非任务打call
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	totalCnt := 0
	for rows.Next() {
		rows.Scan(&totalCnt)
	}

	sql = fmt.Sprintf(`select count(id) as cnt from ddcallForSingers where uin = %d and type in (1,2,3,4,5,6,7)  and ts >= %d and ts <= %d`, uin, minTs, maxTs) //type = 8 为非任务打call
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	cnt := 0
	for rows.Next() {
		rows.Scan(&cnt)
	}

	effictiveCnt := totalCnt - cnt

	ret.Info.LeftCallCnt = Config.NormalCall.Cnt - effictiveCnt
	if ret.Info.LeftCallCnt < 0 {
		ret.Info.LeftCallCnt = 0
	}

	//获取今日任务剩余数

	ret.Info.TotalTaskCnt = 7

	sql = fmt.Sprintf(`select type from ddcallForSingers where status = 0 and uin = %d and singerId = %d and ts >= %d and ts <= %d group by type`, uin, mySingerId, minTs, maxTs)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	calltypes := []int{1, 2, 3, 4, 5, 6, 7}
	callTypesMap := make(map[int]int)

	for rows.Next() {
		var typ int
		rows.Scan(&typ)
		callTypesMap[typ] = 1
	}

	finishedTaskCnt := 0
	for _, typ := range calltypes {
		if _, ok := callTypesMap[typ]; ok {
			finishedTaskCnt++
		}
	}

	ret.Info.FinishedTaskCnt = finishedTaskCnt

	ret.Status = 1
	log.Debugf("end NormalCallForSinger ret:%+v", ret)
	return
}
