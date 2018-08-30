package question

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"github.com/Luxurioust/excelize"
	"net/http"
	"time"
)

type AutoInput2Req struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type AutoInput2Rsp struct {
}

//
func doAutoInputQuestionNoBoardInfo(req *AutoInput2Req, r *http.Request) (rsp *AutoInput2Rsp, err error) {
	//去除首位空白字符
	AutoInputQuestionNoBoardInfo()

	return

}

func AutoInputQuestionNoBoardInfo() (uids []int64, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//v2question表多加了一个字段  (同问sameAskUid)
	stmt, err := inst.Prepare(`insert into no_boardInfo_question(id, boardId, qTitle, qContent, qImgUrls, qType,isAnonymous, qStatus, createTs) 
		values(?, ?, ?, ?, ?, ?, ?, ? , ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	xlsx, err := excelize.OpenFile("/home/work/yplay_svr/etc/no_boardInfo_question.xlsx")
	if err != nil {
		log.Errorf("error:%v", err)
		return
	}
	// Get value from cell by given sheet index and axis.
	//cell := xlsx.GetCellValue("Sheet1", "B2")
	//fmt.Println(cell)
	// Get sheet index.
	//index := xlsx.GetSheetIndex("Sheet1")
	// Get all the rows in a sheet.
	lineInfo := xlsx.GetRows("Sheet1")

	var qTitle, qImgUrls string
	for _, row := range lineInfo {
		log.Debugf("boardId:%s,info:%s,status:%s", row[0], row[1], row[2])
		_, err = stmt.Exec(0, row[0], qTitle, row[1], qImgUrls, 0,row[2], 0, ts)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}

		log.Debugf("input a line :%+v", row)

		//return
	}
	return
}
