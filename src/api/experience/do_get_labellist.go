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

	//AnswerId   int    `schema:"answerId"`
	LabelName string `schema:"labelName"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type LabelInfo struct {
	LabelId   int    `json:"labelId"`
	LabelName string `json:"labelName"`
}
type GetLabelListRsp struct {
	LabelList []*LabelInfo `json:"labelList"`
	TotalCnt  int          `json:"totalCnt"`
}

func doGetLabelList(req *GetLabelListReq, r *http.Request) (rsp *GetLabelListRsp, err error) {

	log.Debugf("uin %d, GetLabelListReq %+v", req.Uin, req)

	labelList, totalCnt, err := GetLabelList(req.Uin, req.LabelName, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetLabelListReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetLabelListRsp{labelList, totalCnt}

	log.Debugf("uin %d, PostLikeRsp succ, %+v", req.Uin, rsp)

	return
}

func GetLabelList(uin int64, labelName string, pageNum, pageSize int) (labelList []*LabelInfo, totalCnt int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	labelList, totalCnt, err = getLabelInfo(inst, labelName, pageNum, pageSize)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	return
}

func getLabelInfo(inst *sql.DB, labelName string, pageNum, pageSize int) (labelList []*LabelInfo, totalCnt int, err error) {

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	sql := fmt.Sprintf(`select labelId,labelName from experience_label  where locate('%s',labelName) limit %d,%d`, labelName, s, e)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	totalCnt = 0
	for rows.Next() {
		var labelInfo LabelInfo
		rows.Scan(&labelInfo.LabelId, &labelInfo.LabelName)

		labelList = append(labelList, &labelInfo)
		totalCnt++
	}
	return
}

func getLabelInfoByBoardId(boardId int, labelName string, pageNum, pageSize int) (labelList []*LabelInfo, totalCnt int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	sql := fmt.Sprintf(`select labelId,labelName from experience_label  where locate('%s',labelName) 
 and labelId in (select labelId from experience_share where boardId = %d) limit %d,%d`, labelName, boardId, s, e)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	totalCnt = 0
	for rows.Next() {
		var labelInfo LabelInfo
		rows.Scan(&labelInfo.LabelId, &labelInfo.LabelName)

		labelList = append(labelList, &labelInfo)
		totalCnt++
	}
	return
}
