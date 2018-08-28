package experience

import (
	"api/board"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

//拉去最新经验贴 只展示名字

type GetExpHomeReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int `schema:"boardId"`
	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type ExpHome struct {
	BoardId   int    `json:"boardId"`
	LabelId   int    `json:"labelId"`
	LabelName string `json:"labelName"`
	Count     int    `json:"count"`
}
type GetExpHomeRsp struct {
	ExpHome   []*ExpHome            `json:"expHome"`
	Operators []*st.UserProfileInfo `json:"operators"`
	IsAdmin   bool                  `json:"isAdmin"`
}

func doGetExpHome(req *GetExpHomeReq, r *http.Request) (rsp *GetExpHomeRsp, err error) {

	log.Debugf("uin %d, GetExpHomeReq succ, %+v", req.Uin, req)

	expHome, operators, isAdmin, err := GetExpHome(req.Uin, req.BoardId)
	if err != nil {
		log.Errorf("uin %d, GetExpHomeReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetExpHomeRsp{expHome, operators, isAdmin}

	log.Debugf("uin %d, GetExpHomeRsp succ, %+v", req.Uin, rsp)

	return
}

func GetExpHome(uin int64, boardId int) (expHome []*ExpHome, operators []*st.UserProfileInfo, isAdmin bool, err error) {
	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	currentTime := time.Now().Unix()

	// boardId labelId labelName startTime endTime cnt

	sql := fmt.Sprintf(`select experience_label.labelId,experience_label.labelName from experience_home,experience_label
							where experience_home.labelId = experience_label.labelId 
							and experience_label.boardId = %d and experience_label.boardId = experience_home.boardId
							and experience_home.startTime <=%d<=experience_home.endTime`, boardId, currentTime)

	log.Debugf("sql:%s\n", sql)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	expHome = make([]*ExpHome, 0)

	for rows.Next() {
		var expHomeTmp ExpHome
		rows.Scan(&expHomeTmp.LabelId, &expHomeTmp.LabelName)
		sql = fmt.Sprintf(`select count(*) as cnt  from experience_share where labelId = %d`, expHomeTmp.LabelId)

		rows2, err2 := inst.Query(sql)
		if err2 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
			log.Error(err2)
			continue
		}
		for rows2.Next() {
			rows2.Scan(&expHomeTmp.Count)
		}
		// count 为0 的经验弹 过滤
		if expHomeTmp.Count == 0 {
			continue
		}
		expHome = append(expHome, &expHomeTmp)
	}

	operators, isAdmin, err = board.GetExpAngelInfoList(uin, boardId)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	return
}
