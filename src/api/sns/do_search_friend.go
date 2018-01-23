package sns

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type SearchFriendReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	UserName string `schema:"userName"`
}

type SearchFriendInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	//	Age        int    `json:"age"`
	Grade      int    `json:"grade"`
	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"schoolName"`
	Phone      string `json:"phone"`

	Status int   `json:"status"` //0非好友 1好友 2已经邀请对方 3对方已经邀请我
	MsgId  int64 `json:"msgId"`  //邀请消息ID
}

type SearchFriendRsp struct {
	Friends []*SearchFriendInfo `json:"friends"`
}

func doSearchFriend(req *SearchFriendReq, r *http.Request) (rsp *SearchFriendRsp, err error) {

	log.Debugf("uin %d, SearchFriendReq %+v", req.Uin, req)

	friends, err := SearchFriends(req.Uin, req.UserName)
	if err != nil {
		log.Errorf("uin %d, SearchFriendRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &SearchFriendRsp{friends}

	log.Debugf("uin %d, SearchFriendRsp succ, %+v", req.Uin, rsp)

	return
}

func SearchFriends(uin int64, userName string) (friends []*SearchFriendInfo, err error) {

	friends = make([]*SearchFriendInfo, 0)

	if uin == 0 || len(userName) == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, phone, nickName, headImgUrl, gender, grade, schoolId, schoolType, schoolName from profiles where userName = ?`)
	rows, err := inst.Query(sql, userName)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var fi SearchFriendInfo
		rows.Scan(&fi.Uin, &fi.Phone, &fi.NickName, &fi.HeadImgUrl, &fi.Gender, &fi.Grade, &fi.SchoolId, &fi.SchoolType, &fi.SchoolName)

		if len(fi.HeadImgUrl) > 0 {
			fi.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", fi.HeadImgUrl)
		}

		friends = append(friends, &fi)
	}

	if len(friends) > 0 {

		if friends[0].Uin == uin {
			friends[0].Status = constant.ENUM_SNS_STATUS_IS_FRIEND
			return
		}

		//对方已经申请加我，返回消息ID
		status, msgId, err1 := st.GetUinStatusWithMe2(uin, friends[0].Uin)
		if err1 != nil {
			return
		}

		friends[0].Status = status
		friends[0].MsgId = msgId
	}

	return
}
