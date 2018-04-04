package activity

import (
	"api/story"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type CallForSingerReq struct {
	Uin   int64  `schema:"uin"` //为爱豆打call的用户
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Type  int    `schema:"type"` //打call类型 1 分享给QQ好友，2 分享到QQ空间，3 分享给微信好友， 4 分享到微信朋友圈， 5 完成两轮答题，6 添加10位好友 ， 7 在个人墙发表一条动态
}

type CallForSingerRsp struct {
	Status    int            `json:"status"`  // 0 其他错误， 1 打call成功，  2 还没有成为任何一位歌手的粉丝 , 3 今日打call类型重复
	Singer    SingerInfo     `json:"singer"`  //爱豆个人信息
	CallInfos []CallTypeInfo `json:callInfos` // 各种打call类型完成情况信息数组
}

type CallTypeInfo struct {
	Type         int `json:"type"`         // 打call类型
	FinishedCnt  int `json:"finishedCnt"`  // 今日完成此类型任务已操作数量
	EffectiveCnt int `josn:"effectiveCnt"` // 今日完成此类型任务已操作中的有效数量
	LeftCnt      int `json:"leftCnt"`      // 今日完成此类型任务剩余操作数量
	Status       int `json:"status"`       // 今日此类型的任务是否已完成的最终状态 0 未完成，1 完成
}

func doCallForSinger(req *CallForSingerReq, r *http.Request) (rsp *CallForSingerRsp, err error) {
	log.Debugf("start doCallForSinger uin:%d  type:%d", req.Uin, req.Type)
	status, err := CallForSinger(req.Uin, req.Type)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	callInfos := make([]CallTypeInfo, 0)
	if status == 1 {
		callInfos, err = getCallTypeInfos(req.Uin)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
	}

	singer, err := getSingerInfo(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &CallForSingerRsp{status, singer, callInfos}
	log.Debugf("end doCallForSinger status:%d, singer:%+v, callInfos:%+v", status, singer, callInfos)
	return
}

func CallForSinger(uin int64, typ int) (status int, err error) {
	log.Debugf("start CallForSinger uin:%d, typ:%d", uin, typ)

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

	singerId := 0
	for rows.Next() {
		rows.Scan(&singerId)
	}

	if singerId == 0 {
		status = 2
		log.Debugf("uin:%d has not be any singer`s fans")
		return
	}

	t := time.Now()
	h, m, s := t.Clock()
	ts := t.Unix()

	minTs := ts - int64(h*3600+m*60+s)
	maxTs := minTs + int64(24*3600)

	sql = fmt.Sprintf(`select type from ddcallForSingers where uin = %d and singerId = %d and type = %d and ts >= %d and ts <= %d`, uin, singerId, typ, minTs, maxTs)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		status = 3
		log.Debugf("uin:%d has called for singer(%d) by type(%d) today", uin, singerId, typ)
		return
	}

	stmt, err := inst.Prepare(`insert into ddcallForSingers values(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts1 := time.Now().Unix()
	stat := 0
	_, err = stmt.Exec(0, singerId, uin, ts1, typ, stat)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	status = 1
	log.Debugf("end CallForSinger")
	return
}

func getCallTypeInfos(uin int64) (callInfos []CallTypeInfo, err error) {
	log.Debugf("start getCallTypeInfos uin:%d", uin)

	callInfos = make([]CallTypeInfo, 0)
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

	singerId := 0
	for rows.Next() {
		rows.Scan(&singerId)
	}

	t := time.Now()
	h, m, s := t.Clock()
	ts := t.Unix()

	minTs := ts - int64(h*3600+m*60+s)
	maxTs := minTs + int64(24*3600)

	sql = fmt.Sprintf(`select type from ddcallForSingers where status = 0 and uin = %d and singerId = %d and ts >= %d and ts <= %d`, uin, singerId, minTs, maxTs)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	callTypeMap := make(map[int]int)
	for rows.Next() {
		var typ int
		rows.Scan(&typ)
		if typ >= 1 && typ <= 7 {
			var info CallTypeInfo
			info.Status = 1
			info.LeftCnt = 0
			info.Type = typ
			callTypeMap[typ] = 1

			switch typ {
			case 1, 2, 3, 4, 7:
				info.EffectiveCnt = 1
				info.FinishedCnt = 1
				break
			case 5:
				info.EffectiveCnt = 24
				info.FinishedCnt = 24
				break
			case 6:
				info.EffectiveCnt = 10
				info.FinishedCnt = 10
				break
			default:
			}

			callInfos = append(callInfos, info)
		} else {
			log.Errorf("wrong call typ:%d", typ)
		}
	}

	totalCallTypes := []int{1, 2, 3, 4, 5, 6, 7}

	unFinishedCallTypes := make([]int, 0)
	for _, k := range totalCallTypes {
		if _, ok := callTypeMap[k]; !ok && k >= 5 && k <= 7 { // 1,2,3,4 是客户端逻辑，5，6，7是服务器逻辑
			unFinishedCallTypes = append(unFinishedCallTypes, k)
		}

		if _, ok := callTypeMap[k]; !ok && k >= 1 && k <= 4 { // 1,2,3,4 是客户端逻辑，5，6，7是服务器逻辑
			var info CallTypeInfo
			info.Status = 0
			info.LeftCnt = 1
			info.EffectiveCnt = 0
			info.FinishedCnt = 0
			info.Type = k

			callInfos = append(callInfos, info)
		}

	}

	if len(unFinishedCallTypes) > 0 {
		ret, err1 := updateCallTypeInfos(uin, unFinishedCallTypes)
		if err1 != nil {
			err = err1
			log.Errorf(err1.Error())
			return
		}
		for _, k := range unFinishedCallTypes {
			if _, ok := ret[k]; ok {
				callInfos = append(callInfos, ret[k])
			}
		}
	}

	log.Debugf("end getCallTypeInfos callInfos:%+v", callInfos)
	return
}

func updateCallTypeInfos(uin int64, callTypes []int) (ret map[int]CallTypeInfo, err error) {
	log.Debugf("start updateCallTypeInfos uin:%d callTypes:%+v", uin, callTypes)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	ret = make(map[int]CallTypeInfo)
	for _, typ := range callTypes {
		var info CallTypeInfo
		info.Status = 0
		info.LeftCnt = 0
		info.EffectiveCnt = 0
		info.FinishedCnt = 0
		info.Type = typ

		t := time.Now()
		h, m, s := t.Clock()
		ts := t.Unix()

		minTs := ts - int64(h*3600+m*60+s)
		maxTs := minTs + int64(24*3600)

		if typ == 5 {

			sql := fmt.Sprintf(`select count(id) as cnt from actRecords where uin = %d and ts >= %d and ts <= %d`, uin, minTs, maxTs)
			rows, err1 := inst.Query(sql)
			if err1 != nil {
				err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
				log.Errorf(err1.Error())
				return
			}
			defer rows.Close()

			for rows.Next() {
				var cnt int
				rows.Scan(&cnt)
				info.FinishedCnt = cnt
				info.EffectiveCnt = cnt
			}
			info.LeftCnt = 24 - info.EffectiveCnt
			if info.LeftCnt == 0 {
				info.Status = 1
			}

		} else if typ == 6 {

			sql := fmt.Sprintf(`select status from addFriendMsg where fromUin = %d and ts >= %d and ts <= %d`, uin, minTs, maxTs)
			rows, err1 := inst.Query(sql)
			if err1 != nil {
				err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
				log.Errorf(err1.Error())
				return
			}
			defer rows.Close()

			for rows.Next() {
				var status int
				rows.Scan(&status)
				if status == 1 {
					info.EffectiveCnt++
				}
				info.FinishedCnt++
			}
			info.LeftCnt = 10 - info.EffectiveCnt

			if info.LeftCnt == 0 {
				info.Status = 1
			}

		} else if typ == 7 {
			cnt, err1 := story.GetMyStoriesCnt(uin)
			if err1 != nil {
				log.Errorf(err1.Error())
				return
			}
			if cnt > 0 {
				info.Status = 1
				info.EffectiveCnt = 1
				info.FinishedCnt = 1
			}
		}

		ret[typ] = info
		if info.Status == 1 {
			_, err = CallForSinger(uin, typ)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
		}
	}
	log.Debugf("end updateCallTypeInfos ret:%+v", ret)
	return
}

//获取爱豆个人信息
func getSingerInfo(uin int64) (singer SingerInfo, err error) {
	log.Debugf("start getSingerInfo uin:%d", uin)

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

	singerId := 0
	for rows.Next() {
		rows.Scan(&singerId)
	}

	sql = fmt.Sprintf(`select singerId, uin, nickName, headImgUrl, gender, deptName, grade from ddsingers where status = 0 and singerId = %d`, singerId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&singer.SingerId, &singer.Uin, &singer.NickName, &singer.HeadImgUrl, &singer.Gender, &singer.DeptName, &singer.Grade)
	}

	log.Debugf("end getSingerInfo singer:%+v", singer)
	return
}
