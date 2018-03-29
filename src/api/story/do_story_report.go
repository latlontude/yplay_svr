package story

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
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

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	stmt, err := inst.Prepare(`insert into report values(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, storyId, typ, desc, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	log.Debugf("end StoryReport")
	return
}
