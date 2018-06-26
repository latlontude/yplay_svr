package story

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type StoryReportReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	StoryId int64  `schema:"storyId"`
	Type    int    `schema:"type"` //举报类型
	Desc    string `schema:"desc"` // 举报文本
}

type StoryReportRsp struct {
}

func doStoryReport(req *StoryReportReq, r *http.Request) (rsp *StoryReportRsp, err error) {

	log.Debugf("uin %d, doStoryReport %+v", req.Uin, req)

	err = StoryReport(req.Uin, req.StoryId, req.Type, req.Desc)
	if err != nil {
		log.Errorf("uin %d, StoryReportRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &StoryReportRsp{}

	log.Debugf("uin %d, StoryReportRsp succ, %+v", req.Uin, rsp)

	return
}

func StoryReport(uin, storyId int64, typ int, desc string) (err error) {
	log.Debugf("start StoryReport uin:%d storyId:%d typ:%d desc:%s", uin, storyId, typ, desc)

	if typ < 1 || typ > 5 {
		err = rest.NewAPIError(constant.E_INVALID_REPORT_TYPE, "report type invalid")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select id from report where status = 0 and uin = %d and storyId = %d`, uin, storyId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rest.NewAPIError(constant.E_REPORT_TYPE_REPEAT, "repeat report")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_STORY_MSG)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", storyId)
	storyMsg, err := app.Get(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("storyMsg:%s", storyMsg)

	var si st.StoryInfo
	err = json.Unmarshal([]byte(storyMsg), &si)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	stmt, err := inst.Prepare(`insert into report values(?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, si.Uin, storyId, storyMsg, typ, desc, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	log.Debugf("end StoryReport")
	return
}

func RecordStory(uin, storyId int64, typ int, data, text, thumbnailImgUrl string) (err error) {
	log.Debugf("start RecordStory uin:%d, storyId:%d, typ:%d, data:%s, text:%s, humbnailImgUrl:%s", uin, storyId, typ, data, text, thumbnailImgUrl)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	stmt, err := inst.Prepare(`insert into story values(?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, storyId, typ, data, text, thumbnailImgUrl, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end RecordStory")
	return
}
