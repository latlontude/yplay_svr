package question

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AutoQuestionReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	BoardId int    `schema:"boardId"`
}

type AutoQuestionRsp struct {
	ids      []int   `json:"ids"`
	Qids     []int64 `json:"qids"`
	RestTime int64   `json:"restTime"` //剩余时间
	Msg      string  `json:"msg"`
}

//
func doAutoQuestion(req *AutoQuestionReq, r *http.Request) (rsp *AutoQuestionRsp, err error) {
	//去除首位空白字符
	ids, qids, restTime, msg, err := AutoQuestion(req.BoardId)

	if err != nil {

	}

	rsp = &AutoQuestionRsp{ids, qids, restTime, msg}

	log.Debugf("AutoQuestionRsp : %+v", rsp)

	return

}

func AutoQuestion(boardId int) (ids []int, qids []int64, restTime int64, msg string, err error) {

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_STAT)
	if err != nil {
		log.Errorf(err.Error())
		//return
	}

	key := fmt.Sprintf("auto_question_%d", boardId)
	storeTime, err := app.Get(key)

	if err != nil {

	}

	now := time.Now().Unix()
	tmp, _ := strconv.ParseInt(storeTime, 10, 64)
	restTime = now - tmp

	if restTime < 300 {
		msg = "please wait 5 min"
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	auto_question_cnt := 0
	sql := fmt.Sprintf(`select id,boardId,qTitle,qContent,qImgUrls,qType,isAnonymous from auto_question where qStatus = 0 and boardId = %d limit 5`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	defer rows.Close()

	ids = make([]int, 0)
	qids = make([]int64, 0)

	index := 0

	randomUids, err := GetRandomUids(boardId)
	if err != nil {
		log.Errorf("query random uids error")
	}

	randomLength := len(randomUids)

	for rows.Next() {
		var id int
		var boardId int
		var qTitle string
		var qContent string
		var qImgUrls string
		var qType int
		var isAnonymous bool
		var ext string
		rows.Scan(&id, &boardId, &qTitle, &qContent, &qImgUrls,&qType, &isAnonymous)
		qid, err := PostQuestion(randomUids[index], boardId, qTitle, strings.Trim(qContent, " \n\t"), qImgUrls, qType,isAnonymous, ext)
		time.Sleep(time.Duration(1)*time.Second)
		if err != nil {

		}
		ids = append(ids, id)
		qids = append(qids, qid)

		index++
		if index > randomLength {
			index = 0
		}
		auto_question_cnt++
	}

	//题库没有题了 直接返回
	if auto_question_cnt == 0 {
		msg = "rest question is zero"
		return
	}

	var idstr string
	for i, id := range ids {
		if i != len(ids)-1 {
			idstr += fmt.Sprintf(`%d,`, id)
		} else {
			idstr += fmt.Sprintf(`%d`, id)
		}
	}

	log.Debugf("test ids = %s", idstr)
	sql = fmt.Sprintf(`update auto_question set qStatus = 1  where id in (%s) and boardId = %d `, idstr, boardId)
	_, err = inst.Query(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	value := fmt.Sprintf("%d", now)
	err = app.Set(key, value)

	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg = "success"

	return
}

func GetRandomUids(boardId int) (uids []int64, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select uin from profiles where locate('1406666',phone) and phone not in (14066660301,14066660320,14066660353,18866668888) and 
schoolId in (select schoolId from v2boards where boardId = %d)  ORDER BY RAND() limit 5`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	uids = make([]int64, 0)
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		uids = append(uids, uid)
	}

	log.Debugf("uids:%+v", uids)
	return
}
