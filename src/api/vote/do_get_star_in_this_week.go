package vote

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type GetStarReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetStarRsp struct {
	Uin        int64  `schema:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	SchoolName string `json:"schoolName"`
	SchoolType int    `json:"schoolType"`
	Grade      int    `json:"grade"`
	Cnt        int    `json:"cnt"`        // 连续几周成为排行榜第一名
	DiamondSet []int  `josn:"diamondSet"` // 每周获得的钻石总数列表
}

func doGetStarInThisWeek(req *GetStarReq, r *http.Request) (rsp *GetStarRsp, err error) {

	log.Debugf("uin %d, GetStarInThisWeek %+v", req.Uin, req)

	info, err := GetStarOfWeek(req.Uin, 0)
	if err != nil {
		log.Errorf("uin %d, GetStarInThisWeek error, %s", req.Uin, err.Error())
		return
	}

	last := 1
	for {
		ret, _ := GetStarOfWeek(req.Uin, last)

		if ret.Uin == info.Uin {
			info.Cnt++
			info.DiamondSet = append(info.DiamondSet, ret.DiamondSet...)
			last++
		} else {
			break
		}
	}

	rsp = &info

	log.Debugf("uin %d, GetStarInThisWeek succ, %+v", req.Uin, rsp)

	return
}

func GetStarOfWeek(uin int64, last int) (ret GetStarRsp, err error) {

	log.Errorf("start GetStarInThisWeek")
	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	s := GetFirstDayOfThisWeek()
	e := time.Now().Unix()

	if last != 0 {
		s = s - int64(last*7*24*3600)
		e = s + int64(7*24*3600)
	}

	log.Errorf("start time:%s end time:%s", time.Unix(s, 0).Format("2006-01-02 15:04:05 PM"), time.Unix(e, 0).Format("2006-01-02 15:04:05 PM"))

	//要全部查询出来 然后找同校同年级的
	sql := fmt.Sprintf(`select voteToUin, count(id) as cnt from voteRecords where ts > %d and ts < %d group by voteToUin order by cnt desc`, s, e)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	in := false
	uidsSlice := make([]int64, 0)
	uidCntMap := make(map[int64]int)
	for rows.Next() {
		var uid int64
		var cnt int

		rows.Scan(&uid, &cnt)

		uidCntMap[uid] = cnt
		uidsSlice = append(uidsSlice, uid)

		if uid == uin {
			in = true
		}
	}

	uidsSlice = append(uidsSlice, uin) //也要获取本用户的信息，用于后续查找同校同年级的人

	log.Errorf("uidsSlice:%+v", uidsSlice)

	res, err := st.BatchGetUserProfileInfo(uidsSlice)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var tmpRet GetStarRsp
	for _, uid := range uidsSlice {
		if _, ok := res[uid]; ok {
			if res[uin].SchoolId == res[uid].SchoolId && res[uin].Grade == res[uid].Grade {
				tmpRet.Uin = uid
				tmpRet.NickName = res[uid].NickName
				tmpRet.HeadImgUrl = res[uid].HeadImgUrl
				tmpRet.SchoolName = res[uid].SchoolName
				tmpRet.SchoolType = res[uid].SchoolType
				tmpRet.Grade = res[uid].Grade
				tmpRet.Cnt = 1
				tmpRet.DiamondSet = append(tmpRet.DiamondSet, uidCntMap[uid])
				break

			}
		}
	}
	log.Errorf("star:%+v", tmpRet)

	if tmpRet.Uin == uin {
		if in {
			ret = tmpRet
		}
	} else {
		ret = tmpRet
	}

	log.Errorf("end GetStarInThisWeek")
	return
}

func GetFirstDayOfThisWeek() (ts int64) {

	t := time.Now()
	week := int(t.Weekday())
	h, m, s := t.Clock()

	ts = t.Unix()

	ts -= int64((week-1)*24*3600 + h*3600 + m*60 + s)
	return

}
