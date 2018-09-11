package board

import (
	"api/common"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type GetBoardsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetBoardsRsp struct {
	Boards []*st.BoardInfo `json:"boards"` //选项
}

func doGetBoards(req *GetBoardsReq, r *http.Request) (rsp *GetBoardsRsp, err error) {

	log.Debugf("uin %d, GetBoardsReq %+v", req.Uin, req)

	boards, err := GetBoards(req.Uin)

	if err != nil {
		log.Errorf("uin %d, GetBoards error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetBoardsRsp{boards}

	log.Debugf("uin %d, GetBoardsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetBoards(uin int64) (boards []*st.BoardInfo, err error) {

	log.Debugf("start GetBoards uin:%d", uin)

	boards = make([]*st.BoardInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	uInfo, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}
	sql := fmt.Sprintf(`select boardId, boardName, boardIntro, boardIconUrl, boardStatus, schoolId, ownerUid, createTs from v2boards where schoolId = %d and boardStatus = 0`, uInfo.SchoolId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	//判断有没有墙
	hasBoardId := false

	for rows.Next() {
		var info st.BoardInfo
		var uid int64

		rows.Scan(
			&info.BoardId,
			&info.BoardName,
			&info.BoardIntro,
			&info.BoardIconUrl,
			&info.BoardStatus,
			&info.SchoolId,
			&uid,
			&info.CreateTs)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		} else {
			//没有墙主信息 构造一个空对象
			info.OwnerInfo = &st.UserProfileInfo{}
		}

		if info.SchoolId > 0 {
			si, err1 := st.GetSchoolInfo(info.SchoolId)
			if err != nil {
				log.Error(err1.Error())
				continue
			}

			info.SchoolName = si.SchoolName
			info.SchoolType = si.SchoolType
		}

		followCnt, _ := GetFollowCnt(info.BoardId)
		info.FollowCnt = followCnt

		//boardInfo 返回 是否是经验弹管理员
		isAdmin, err2 := common.CheckPermit(uin, info.BoardId, 0)
		if err2 != nil {
			log.Error(err2.Error())
		}
		info.IsAdmin = isAdmin

		log.Debugf("isAdmin : %v", isAdmin)

		boards = append(boards, &info)

		hasBoardId = true
	}
	//该学校还没有开墙 插入一条墙记录  并且学校存在
	if !hasBoardId && uInfo.SchoolId < 9999997 {
		//不用 :=  局部变量覆盖boards
		boards, err = CreateBoardInfo(uin, uInfo)
		if err != nil {
			log.Debugf("err1:%v", err)
		}
		uinList, err2 := AutoRegister(uInfo)
		if err2 != nil {
			log.Debugf("err2:%v", err2)
		}
		//创建完墙 插入5问题
		err3 := InsertQuestions(boards[0].BoardId, uinList)

		if err3 != nil {
			log.Debugf("err3:%v", err3)
		}
	}
	return
}

func CreateBoardInfo(uin int64, uInfo *st.UserProfileInfo) (boards []*st.BoardInfo, err error) {
	boards = make([]*st.BoardInfo, 0)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	//3.插入表
	stmt, err := inst.Prepare(`insert into v2boards values(?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()
	createTs := time.Now().Unix()
	var boardIconUrl string
	res, err := stmt.Exec(0, uInfo.SchoolName, uInfo.SchoolName, boardIconUrl, 0, uInfo.SchoolId, 0, createTs)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	//获取新增数据id
	boardId, err := res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}
	var info st.BoardInfo

	info.BoardId = int(boardId)
	info.BoardName = uInfo.SchoolName
	info.BoardIntro = uInfo.SchoolName
	info.BoardStatus = 0
	info.SchoolId = uInfo.SchoolId
	info.OwnerInfo = &st.UserProfileInfo{}
	info.CreateTs = int(createTs)

	if info.SchoolId > 0 {
		si, err1 := st.GetSchoolInfo(info.SchoolId)
		if err != nil {
			log.Error(err1.Error())
		}
		info.SchoolName = si.SchoolName
		info.SchoolType = si.SchoolType
	}

	followCnt, _ := GetFollowCnt(info.BoardId)
	info.FollowCnt = followCnt

	//boardInfo 返回 是否是经验弹管理员
	isAdmin, err2 := common.CheckPermit(uin, info.BoardId, 0)
	if err2 != nil {
		log.Error(err2.Error())
	}
	info.IsAdmin = isAdmin
	boards = append(boards, &info)
	log.Debugf("ceate boards :%v", boards)
	return
}

func InsertQuestions(boardId int, registerUin []int64) (err error) {

	log.Debugf("insertNewQuestions")
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)

	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select qTitle,qContent,qImgUrls,qType,isAnonymous from no_boardInfo_question `)
	log.Debugf("insertNewQuestions sql:%s", sql)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()
	sqlArr := make([]string, 0)
	index := 0
	registerUinIndex := 0

	log.Debugf("registerUin:%v", registerUin)
	for rows.Next() {
		var info st.V2QuestionInfo
		var qStatus int
		var sameAskUid string
		rows.Scan(
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.QType,
			&info.IsAnonymous)

		log.Debugf("info:%v", info)

		if registerUinIndex == len(registerUin) {
			registerUinIndex = 0
		}
		createTs := int(time.Now().Unix()) + index
		sqlArr = append(sqlArr, fmt.Sprintf(`insert into v2questions values(%d, %d, %d, '%s', '%s', '%s', %d, %t, %d, %d, %d, '%s' ,'%s')`,
			0, boardId, registerUin[registerUinIndex], info.QTitle, info.QContent, info.QImgUrls, info.QType, info.IsAnonymous, qStatus, createTs, info.ModTs, sameAskUid, info.Ext))

		registerUinIndex++
		index++
	}

	log.Debugf("sqlArr:%v", sqlArr)

	if len(sqlArr) == 0 {
		log.Debugf("no question in no_boardInfo_question")
		return
	}
	err = mydb.Exec(inst, sqlArr)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	return
}

//
//func getFollowCnt(boardId int) (cnt int, err error) {
//	//log.Debugf("start getFollowCnt boardId:%d", boardId)
//
//	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
//	if inst == nil {
//		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
//		log.Error(err)
//		return
//	}
//
//	sql := fmt.Sprintf(`select count(id) as cnt from v2follow where boardId = %d and status = 0`, boardId)
//
//	rows, err := inst.Query(sql)
//	if err != nil {
//		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
//		log.Error(err)
//		return
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//		rows.Scan(&cnt)
//	}
//
//	//log.Debugf("end getFollowCnt boardId:%d cnt:%d", boardId, cnt)
//	return
//}

func GetFollowCnt(boardId int) (followCnt int, err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(uin) as cnt from profiles where schoolId in (select schoolId  from v2boards where boardId = %d ) `, boardId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&followCnt)
	}
	log.Debugf("board followCnt :%d", followCnt)

	return
}

func GetBoardInfoByBoardId(uin int64, boardId int) (info st.BoardInfo, err error) {

	//TODO: 缓存boardInfo 不需要每个问题都要拉去boardInfo
	boardInfo, ok := boardMap[boardId]
	if ok {
		info = *boardInfo
		log.Debugf("board in map，boardId:%d", boardId)
		return
	} else {
		log.Debugf("board not in map boardId:%d", boardId)
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select boardId, boardName, boardIntro, boardIconUrl, boardStatus, schoolId, ownerUid, createTs from v2boards where boardId = %d and boardStatus = 0`, boardId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64

		rows.Scan(
			&info.BoardId,
			&info.BoardName,
			&info.BoardIntro,
			&info.BoardIconUrl,
			&info.BoardStatus,
			&info.SchoolId,
			&uid,
			&info.CreateTs)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		if info.SchoolId > 0 {
			si, err1 := st.GetSchoolInfo(info.SchoolId)
			if err != nil {
				log.Error(err1.Error())
				continue
			}

			info.SchoolName = si.SchoolName
			info.SchoolType = si.SchoolType
		}

		follwCnt, _ := GetFollowCnt(info.BoardId)
		info.FollowCnt = follwCnt

		//boardInfo 返回 是否是经验弹管理员
		isAdmin, err2 := common.CheckPermit(uin, info.BoardId, 0)
		if err2 != nil {
			log.Error(err2.Error())
		}
		info.IsAdmin = isAdmin

		boardMap[boardId] = &info
	}

	//log.Debugf("end GetBoards boards:%+v", info)
	return
}
