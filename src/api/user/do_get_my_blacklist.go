package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type GetMyBlacklistReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetMyBlacklistRsp struct {
	Users []*st.UserProfileInfo `json:"users"`
}

func doGetMyBlacklist(req *GetMyBlacklistReq, r *http.Request) (rsp *GetMyBlacklistRsp, err error) {

	log.Debugf("uin %d, doGetMyBlacklist", req.Uin)

	users, err := GetMyBlacklist(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetMyBlacklistRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetMyBlacklistRsp{users}

	log.Debugf("uin %d, GetMyBlacklistRsp succ, %+v", req.Uin, rsp)

	return
}

func GetMyBlacklist(uin int64) (users []*st.UserProfileInfo, err error) {
	log.Debugf("start GetMyBlacklist uin:%d ", uin)

	users = make([]*st.UserProfileInfo, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	uids := make([]int64, 0)
	sql := fmt.Sprintf(`select uid from pullBlackUser where uin = %d`, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		uids = append(uids, uid)
	}

	ret, err := st.BatchGetUserProfileInfo(uids)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, info := range ret {
		users = append(users, info)
	}

	log.Debugf("end GetMyBlacklist users:%+v", users)
	return
}
