package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"time"
)

type AddLabelReq struct {
	Uin         int64   `schema:"uin"`
	Token       string  `schema:"token"`
	Ver         int     `schema:"ver"`

	LableName   string  `schema:"labelName"`

}

type AddLabelRsp struct {
	LabelId  int64 `json:"labelId"`
}



func doAddLabel(req *AddLabelReq, r *http.Request) (rsp *AddLabelRsp, err error) {

	log.Debugf("uin %d, AddLabelReq succ, %+v", req.Uin, rsp)

	labelId , err  := AddLabel(req.Uin,req.LableName)

	if err != nil {
		log.Errorf("uin %d, AddLabelReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddLabelRsp{labelId}
	return
}


func AddLabel(uin int64, labelName string) (labelId int64 ,err error){

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into experience_label(labelid, labelName, ownerUid, createtTs) values(?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	res , err := stmt.Exec(0, labelName, uin , ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	labelId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}