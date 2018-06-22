package board

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
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

		follwCnt, _ := getFollowCnt(info.BoardId)
		info.FollowCnt = follwCnt

		boards = append(boards, &info)
	}

	log.Debugf("end GetBoards boards:%+v", boards)
	return
}

func getFollowCnt(boardId int64) (cnt int, err error) {
	log.Debugf("start getFollowCnt boardId:%d", boardId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select count(id) as cnt from v2follow where boardId = %d and status = 0`, boardId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&cnt)
	}

	log.Debugf("end getFollowCnt boardId:%d cnt:%d", boardId, cnt)
	return
}