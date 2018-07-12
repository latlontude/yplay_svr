package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"database/sql"
	"fmt"
	"net/http"
)

type GetLabelListReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid   int    `schema:"qid"`
	LabelName string `schema:"labelName"`

}

type GetLabelListRsp struct {
	LabelList []string `json:"labelList"`
	IsVisible bool      `json:"isVisable"`      //标签是否能选中
}

func doGetLabelList(req *GetLabelListReq, r *http.Request) (rsp *GetLabelListRsp, err error) {

	log.Debugf("uin %d, GetLabelListReq %+v", req.Uin, req)

	labelList,isVisible,err := GetLabelList(req.Uin,req.Qid,req.LabelName)

	if err != nil {
		log.Errorf("uin %d, GetLabelListReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetLabelListRsp{labelList,isVisible}

	log.Debugf("uin %d, PostLikeRsp succ, %+v", req.Uin, rsp)

	return
}

func GetLabelList(uin int64 , qid int , labelName string) (labelList []string,isVisible bool , err error) {


	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	labelList , err = getLabelName(inst,labelName)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	isVisible , err = isVisibleLabel(inst,qid)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	return
}

func getLabelName(inst *sql.DB,labelName string) (labelList []string , err error){

	sql := fmt.Sprintf(`select labelName from experience_label  where locate('%s',labelName);`,labelName)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var labelName string
		rows.Scan(&labelName)
		labelList = append(labelList,labelName)
	}
	return
}


func isVisibleLabel(inst *sql.DB, qid int) (isVisible bool , err error){

	sql := fmt.Sprintf(`select count(qid) from experience_share where qid = %d`,qid)

	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	var qidLabelCount int
	for rows.Next() {
		rows.Scan(&qidLabelCount)
	}

	if qidLabelCount >= constant.EXPERIENCE_QID_LABEL_COUNT {
		isVisible = false
	}

	isVisible = true

	return
}