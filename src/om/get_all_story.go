package om

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

type storyInfo struct {
	OwnerInfo       *st.UserProfileInfo `json:"ownerInfo"`
	OwnerUin        int64               `json:"ownerUin"`
	MsgId           int64               `json:"msgId"`
	MsgType         int                 `json:"msgType"`
	MsgData         string              `json:"msgData"`
	MsgContent      string              `json:"msgContent"`
	ThumbnailImgUrl string              `json:"thumbnailImgUrl"`
	MsgTs           int                 `json:"msgTs"`
}

type GetAllStoryReq struct {
	Params map[string]int `schema:"params"`
}

type GetAllStoryRsp struct {
	Total  int         `json:"total"`
	Storys []storyInfo `json:"storys"`
}

func doGetAllStorys(req *GetAllStoryReq) (rsp GetAllStoryRsp, err error) {
	log.Debugf("start doGetAllStorys params:%+v", req.Params)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	pageNum := req.Params["pageNum"]
	pageSize := req.Params["pageSize"]
	uin := 0
	msgType := 0
	tsStart := 0
	tsEnd := 0

	if _, ok := req.Params["uin"]; ok {
		uin = req.Params["uin"]
	}
	if _, ok := req.Params["msgType"]; ok {
		msgType = req.Params["msgType"]
	}
	if _, ok := req.Params["tsStart"]; ok {
		tsStart = req.Params["tsStart"]
	}
	if _, ok := req.Params["tsEnd"]; ok {
		tsEnd = req.Params["tsEnd"]
	}

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	condition := ""
	if uin != 0 {
		condition += fmt.Sprintf(" where uin = %d ", uin)
	}

	if msgType != 0 {
		if len(condition) == 0 {
			condition += fmt.Sprintf(" where msgType = %d ", msgType)
		} else {
			condition += fmt.Sprintf(" and msgType = %d ", msgType)
		}
	}

	if tsStart != 0 {
		if len(condition) == 0 {
			condition += fmt.Sprintf(" where ts >= %d ", tsStart)
		} else {
			condition += fmt.Sprintf(" and ts >= %d ", tsStart)
		}
	}

	if tsEnd != 0 {
		if len(condition) == 0 {
			condition += fmt.Sprintf(" where ts <= %d ", tsEnd)
		} else {
			condition += fmt.Sprintf(" and ts <= %d ", tsEnd)
		}
	}

	if len(Config.BlackList.Uins) != 0 {
		if len(condition) == 0 {
			condition += fmt.Sprintf("where uin not in (%s)", Config.BlackList.Uins)
		} else {
			condition += fmt.Sprintf("and uin not in (%s)", Config.BlackList.Uins)
		}
	}

	sql := fmt.Sprintf("select count(id) as cnt from story %s", condition)
	log.Debugf("query total sql:%s", sql)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&rsp.Total)
	}

	if rsp.Total == 0 {
		return
	}

	sql = fmt.Sprintf("select uin, msgId, msgType, msgData, msgContent, thumbnailImgUrl, ts from story %s order by ts desc limit %d, %d", condition, s, e)
	log.Debugf("query info sql:%s", sql)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	uids := make([]int64, 0)
	for rows.Next() {

		var info storyInfo
		rows.Scan(&info.OwnerUin, &info.MsgId, &info.MsgType, &info.MsgData, &info.MsgContent, &info.ThumbnailImgUrl, &info.MsgTs)

		rsp.Storys = append(rsp.Storys, info)
		uids = append(uids, info.OwnerUin)
	}

	res, err := st.BatchGetUserProfileInfo(uids)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for i, info := range rsp.Storys {
		if _, ok := res[info.OwnerUin]; ok {
			rsp.Storys[i].OwnerInfo = res[info.OwnerUin]
		}
	}

	log.Debugf("end storeUserInfo")
	return
}
