package label

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"database/sql"
	"fmt"
)

type LabelInfo struct {
	BoardId   int    `json:"boardId"`
	LabelId   int    `json:"labelId"`
	LabelName string `json:"labelName"`
}

func GetLabelList(uin int64, boardId int, labelName string, pageNum, pageSize int) (labelList []*LabelInfo, totalCnt int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	labelList, totalCnt, err = GetLabelInfo(inst, boardId, labelName, pageNum, pageSize)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	return
}

func GetLabelInfo(inst *sql.DB, boardId int, labelName string, pageNum, pageSize int) (labelList []*LabelInfo, totalCnt int, err error) {

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	//按照创建时间递减排序
	sql := fmt.Sprintf(`select labelId,labelName from experience_label  
			where locate('%s',labelName) and boardId = %d order by createTs desc limit %d,%d`, labelName, boardId, s, e)
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

func GetLabelInfoByBoardId(boardId int, labelName string, pageNum, pageSize int) (labelList []*LabelInfo, totalCnt int, err error) {

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
