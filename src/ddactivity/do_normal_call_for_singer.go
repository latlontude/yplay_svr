package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type NormalCallForSingerReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type NormalCallForSingerRsp struct {
	Status int `json:"status"` //0 其他错误， 1 成功, 2 今日打call次数超限, 3 用户还未投票，不能打call
}

func doNormalCallForSinger(req *NormalCallForSingerReq, r *http.Request) (rsp *NormalCallForSingerRsp, err error) {
	log.Debugf("start doNormalCallForSinger uin:%d", req.Uin)
	status, err := NormalCallForSinger(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &NormalCallForSingerRsp{status}
	log.Debugf("end doNormalCallForSinger rsp:%+v", rsp)
	return
}

func NormalCallForSinger(uin int64) (status int, err error) {
	log.Debugf("start NormalCallForSinger uin:%d", uin)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	t := time.Now()
	h, m, s := t.Clock()
	ts := t.Unix()

	minTs := ts - int64(h*3600+m*60+s)
	maxTs := minTs + int64(24*3600)

	sql := fmt.Sprintf(`select count(id) as cnt from ddcallForSingers where uin = %d and type = 8 and ts >= %d and ts <= %d`, uin, minTs, maxTs) //type = 8 为非任务打call
	rows, err := inst.Query(sql)
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

	effticeCnt := totalCnt - cnt

	overrun := false

	if effticeCnt >= Config.NormalCall.Cnt {
		overrun = true
	}

	if overrun {
		status = 2
		log.Debugf("uin:%d nomal call cnt > %d", uin, Config.NormalCall.Cnt)
		return
	}

	sql = fmt.Sprintf(`select singerId from ddsingerFansFromPupu where status = 0 and uin = %d`, uin)
	rows, err = inst.Query(sql)
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
		status = 3
		log.Debugf("uin:%d has not be any singer`s fans")
		return
	}

	stmt, err := inst.Prepare(`insert into ddcallForSingers values(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts = time.Now().Unix()
	stat := 0
	typ := 8
	_, err = stmt.Exec(0, singerId, uin, typ, stat, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	status = 1
	log.Debugf("end NormalCallForSinger")
	return
}
