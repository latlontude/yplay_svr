package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type AddLabelReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId   int    `schema:"boardId"`
	LabelName string `schema:"labelName"`
}

type AddLabelRsp struct {
	LabelId int `json:"labelId"`
}

func doAddLabel(req *AddLabelReq, r *http.Request) (rsp *AddLabelRsp, err error) {

	log.Debugf("uin %d, AddLabelReq succ, %+v", req.Uin, req)

	labelId, err := AddLabel(req.Uin, req.LabelName, req.BoardId)

	if err != nil {
		log.Errorf("uin %d, AddLabelReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddLabelRsp{labelId}
	return
}

func AddLabel(uin int64, labelName string, boardId int) (labelId int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select labelId from experience_label where labelName = '%s' and boardId = %d`, labelName, boardId)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	index := 0
	for rows.Next() {
		rows.Scan(&index)
	}

	if index > 0 {
		labelId = index
		//已经有该标签  不需要添加 直接返回
		log.Debug("repeat add , labelId :%d ", labelId)
		return
	}

	stmt, err := inst.Prepare(`insert into experience_label(labelId, labelName, ownerUid, createTs,boardId) values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	res, err := stmt.Exec(0, labelName, uin, ts, boardId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	labelId = int(lastId)

	return
}
