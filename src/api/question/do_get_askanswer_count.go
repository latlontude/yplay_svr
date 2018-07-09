package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	_ "time"
)

type AskAndAnswerCountReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	Begin int64  `schema:"begin"`
	End   int64  `schema:"end"`
}

type AskAndAnswerCountRsp struct {
	AskPersonNum    int    `json:"askPersonNum"`
	AskPersonArr    string `json:"askPersonArr"` //真实用户
	AskCount        int    `json:"askCount"`
	AnswerPersonNum int    `json:"answerPersonNum"`
	AnswerPersonArr string `json:"answerPersonArr"` //真实用户
	AnswerCount     int    `json:"answerCount"`
}

func doDailyCount(req *AskAndAnswerCountReq, r *http.Request) (rsp *AskAndAnswerCountRsp, err error) {
	log.Debugf("uin %d, AskAndAnswerCountReq %+v", req.Uin, req)

	askPersonNum, askPersonArr, askCount, answerPersonNum, answerPersonArr, answerCount, err := dailyCount(req.Uin, req.Begin, req.End)
	if err != nil {
		log.Errorf("uin %d, AskAndAnswerCountRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AskAndAnswerCountRsp{askPersonNum, askPersonArr, askCount, answerPersonNum, answerPersonArr, answerCount}

	log.Debugf("uin %d, AskAndAnswerCountRsp succ, %+v", req.Uin, rsp)
	return
}

func dailyCount(uin int64, begin int64, end int64) (askPersonNum int, askPersonArr string, askCount int, answerPersonNum int, answerPersonArr string, answerCount int, err error) {

	log.Debugf("dailyCount uin = %d", uin)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//timeStr := time.Now().Format("2006-01-02")
	//t2, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)

	//today
	//todayZeroTime := t2.Unix()
	//tomorrowZeroTime := t2.AddDate(0, 0, 1).Unix()

	//查询运营 uids
	sql := fmt.Sprintf(`	select uin from profiles where locate('1406666',phone) or  locate('1706666',phone) 
				or phone in (15319911212 ,13310875517 ,13823798053,15773237801,13003273787,13656943669,13564556649,
				13665694369,13655843699,13566335576,13569664699,13656636499,18829343422)`)
	rows, err := inst.Query(sql)
	errMsg(err)

	defer rows.Close()

	uids := make([]int64, 0)
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		uids = append(uids, uid)
	}
	str := ""
	for i, uid := range uids {
		if i != len(uids)-1 {
			str += fmt.Sprintf(`%d,`, uid)
		} else {
			str += fmt.Sprintf(`%d`, uid)
		}
	}

	log.Debugf("test uids = %s", str)

	//真实用户提问总数
	sql = fmt.Sprintf(`select count(*) from v2questions where createTs >= %d and createTs < %d 
							and ownerUid not in (%s)`, begin, end, str)
	rows, err = inst.Query(sql)
	errMsg(err)

	for rows.Next() {
		rows.Scan(&askCount)
	}

	log.Debugf("askCount:%d", askCount)

	sql = fmt.Sprintf(`select ownerUid from v2questions where createTs >= %d and createTs < %d and ownerUid not in (%s) group by ownerUid`, begin, end, str)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	realUidArr := make([]int64, 0)
	askPersonNum = 0
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		askPersonNum++
		realUidArr = append(realUidArr, uid)
	}

	for i, uid := range realUidArr {
		if i != len(uids)-1 {
			askPersonArr += fmt.Sprintf(`%d,`, uid)
		} else {
			askPersonArr += fmt.Sprintf(`%d`, uid)
		}
	}

	log.Debugf("askPersonNum:%d,askPerson=%s", askPersonNum, askPersonArr)

	//answer

	sql = fmt.Sprintf(`select count(*) from v2answers where answerTs >= %d and answerTs < %d and ownerUid not in (%s)`, begin, end, str)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		rows.Scan(&answerCount)
	}

	log.Debugf("answerCount:%d", answerCount)

	sql = fmt.Sprintf(`select ownerUid from v2answers where answerTs >= %d and answerTs < %d and ownerUid not in (%s) group by ownerUid`, begin, end, str)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	realUidArr = make([]int64, 0)
	answerPersonNum = 0
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		answerPersonNum++
		realUidArr = append(realUidArr, uid)
	}

	for i, uid := range realUidArr {
		if i != len(uids)-1 {
			answerPersonArr += fmt.Sprintf(`%d,`, uid)
		} else {
			answerPersonArr += fmt.Sprintf(`%d`, uid)
		}
	}

	log.Debugf("answerPersonNum:%d,str=%s", answerPersonNum, answerPersonArr)

	return
}

func errMsg(err error) {
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	return
}
