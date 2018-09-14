package board

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"sort"
	"svr/st"
	"time"
)

// 自定义排序
type uInfo []*st.UserProfileInfo

func (I uInfo) Len() int {
	return len(I)
}

func (I uInfo) Less(i, j int) bool {
	return I[i].Src > I[j].Src
}

func (I uInfo) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

type GetAngelInfoReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	SchoolId int    `schema:"schoolId"`
	BoardId  int    `schema:"boardId"`

	ShowType int `schema:"showType"` //展示方式
}

type GetAngelInfoRsp struct {
	AngelList []*st.UserProfileInfo `json:"angelList"`
	IsAdmin   bool                  `json:"isAdmin"`
}

func doGetAngelInfo(req *GetAngelInfoReq, r *http.Request) (rsp *GetAngelInfoRsp, err error) {

	log.Debugf("uin %d, GetAngelInfoReq %+v", req.Uin, req)

	var angelList []*st.UserProfileInfo
	var isAdmin bool
	if req.ShowType == 1 {
		angelList, isAdmin, err = GetExpAngelInfoList(req.Uin, req.BoardId)
	} else {
		angelList, isAdmin, err = GetAngelInfoList(req.Uin, req.BoardId)
	}

	if err != nil {
		return
	}
	rsp = &GetAngelInfoRsp{angelList, isAdmin}
	log.Debugf("uin %d, rsp %+v", req.Uin, rsp)
	return
}

func GetAngelInfoList(uin int64, boardId int) (angelList []*st.UserProfileInfo, isAdmin bool, err error) {

	angelList = make([]*st.UserProfileInfo, 0)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//先判断墙主在不在列表里

	sql := fmt.Sprintf(`select ownerUid from v2boards where boardId = %d`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	//先把墙主放入管理员列表
	var boardUin int64 = 0
	for rows.Next() {
		rows.Scan(&boardUin)
	}

	sql = fmt.Sprintf(`select uin from experience_admin where uin = %d and boardId = %d `, boardUin, boardId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	var adminBoardUin int64 = 0
	for rows.Next() {
		rows.Scan(&adminBoardUin)
	}

	//墙主不在管理员列表 把他加入
	if adminBoardUin == 0 {
		stmt, err1 := inst.Prepare(`insert into experience_admin values(?, ?, ?, ?, ?, ?)`)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_PREPARE, err1.Error())
			log.Error(err1.Error())
			return
		}
		ts := time.Now().Unix()
		_, err = stmt.Exec(0, boardId, 0, boardUin, ts, 1)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
		}
	}

	log.Debugf("boardInfo :%d", boardUin)

	sql = fmt.Sprintf(`select uin ,ts from experience_admin where boardId = %d`, boardId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	now := time.Now().Unix()
	for rows.Next() {
		var uid int64
		var ts int64
		rows.Scan(&uid, &ts)
		log.Debugf("uin:%d", uid)
		if uid <= 0 {
			continue
		}

		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			continue
		}

		//计算天使的在任时间
		if ts > 0 {
			ui.IsAngelDays = int((now - ts) / 86400)
		} else {
			ui.IsAngelDays = 0
		}

		if boardUin == ui.Uin {
			ui.Src = 1
			isAdmin = true
		}
		//自己排序到第一位
		if uin == ui.Uin {
			ui.Src = 2
		}

		querySql := fmt.Sprintf(`select headImgUrl from secondHeadImgUrl where phone = %s`, ui.Phone)

		log.Debugf("querySql:%s", querySql)

		secondImgRows, err2 := inst.Query(querySql)
		if err2 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
			log.Errorf(err.Error())
			continue
		}

		for secondImgRows.Next() {
			var secondImgUrl string
			secondImgRows.Scan(&secondImgUrl)

			log.Debugf("img:%s", secondImgUrl)
			if len(secondImgUrl) > 0 {
				ui.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", secondImgUrl)
			}
		}

		angelList = append(angelList, ui)
	}

	sort.Sort(uInfo(angelList))

	log.Debugf("admin list :%+v", angelList)

	return
}

func GetExpAngelInfoList(uin int64, boardId int) (angelList []*st.UserProfileInfo, isAdmin bool, err error) {

	angelList = make([]*st.UserProfileInfo, 0)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//先判断墙主在不在列表里

	sql := fmt.Sprintf(`select ownerUid from v2boards where boardId = %d`, boardId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	//先把墙主放入管理员列表
	var boardUin int64 = 0
	for rows.Next() {
		rows.Scan(&boardUin)
	}

	if boardUin == 0{
		log.Debugf("没有墙主  uin=%d,boardId=%d",uin,boardId)
		return
	}

	sql = fmt.Sprintf(`select uin from experience_admin where uin = %d and boardId = %d `, boardUin, boardId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	var adminBoardUin int64 = 0
	for rows.Next() {
		rows.Scan(&adminBoardUin)
	}

	//墙主不在管理员列表 把他加入
	if adminBoardUin == 0 {
		stmt, err1 := inst.Prepare(`insert into experience_admin values(?, ?, ?, ?, ?, ?)`)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_PREPARE, err1.Error())
			log.Error(err1.Error())
			return
		}
		ts := time.Now().Unix()
		_, err = stmt.Exec(0, boardId, 0, boardUin, ts, 1)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
		}
	}

	log.Debugf("boardInfo :%d", boardUin)

	sql = fmt.Sprintf(`select uin from experience_admin where boardId = %d`, boardId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		log.Debugf("uin:%d", uid)
		if uid <= 0 {
			continue
		}
		ui, err1 := st.GetUserProfileInfo(uid)
		if err1 != nil {
			continue
		}
		//自己排序到第一位
		if uin == ui.Uin {
			ui.Src = 1
			isAdmin = true
		}
		//sort
		if boardUin == ui.Uin {
			ui.Src = 2
		}
		angelList = append(angelList, ui)
	}
	sort.Sort(uInfo(angelList))
	log.Debugf("exp angel list :%+v", angelList)
	return
}
