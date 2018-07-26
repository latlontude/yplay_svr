package experience

import (
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
	LabelId   int    `json:"labelId"`
	LabelName string `json:"labelName"`
	Count     int    `json:"count"`
}
type GetExpHomeRsp struct {
	ExpHome   []*ExpHome            `json:"expHome"`
	Operators []*st.UserProfileInfo `json:"operators"`
}

func doGetExpHome(req *GetExpHomeReq, r *http.Request) (rsp *GetExpHomeRsp, err error) {

	log.Debugf("uin %d, GetExpHomeReq succ, %+v", req.Uin, req)

	expHome, operators, err := GetExpHome(req.Uin, req.BoardId)
	if err != nil {
		log.Errorf("uin %d, GetExpHomeReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetExpHomeRsp{expHome, operators}

	log.Debugf("uin %d, GetExpHomeRsp succ, %+v", req.Uin, rsp)

	return
}

func GetExpHome(uin int64, boardId int) (expHome []*ExpHome, operators []*st.UserProfileInfo, err error) {
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
							and experience_home.boardId = %d
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

	//operators
	userMap := make(map[int64]st.UserProfileInfo, 0)

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
		expHome = append(expHome, &expHomeTmp)

		//整理过经验弹的人  找到最新时间
		sql = fmt.Sprintf(`select operator from experience_share where boardId = %d and labelId = %d  group by operator order by ts`, boardId, expHomeTmp.LabelId)
		rows3, err3 := inst.Query(sql)
		if err3 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err3.Error())
			log.Error(err)
			return
		}
		for rows3.Next() {
			var uid int64
			rows3.Scan(&uid)
			//去重
			if _, ok := userMap[uid]; ok {
				log.Debugf("map [k:%d,v:%+v]", uid, userMap[uid])
				continue
			}
			if uid > 0 {
				ui, err1 := st.GetUserProfileInfo(uid)
				if err1 != nil {
					log.Error(err1.Error())
					continue
				}
				userMap[uid] = *ui
				operators = append(operators, ui)
			}
		}

	}
	return
}
