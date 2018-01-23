package sns

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	//"time"
	"strconv"
	"svr/st"
)

type GetAddFriendMsgReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`

	UpdateLastReadMsgId int `schema:"updateLastReadMsgId"`
}

type AddFriendMsgInfo struct {
	MsgId          int64  `json:"msgId"`
	Uin            int64  `json:"uin"`
	FromUin        int64  `json:"fromUin"`
	FromNickName   string `json:"fromNickName"`
	FromHeadImgUrl string `json:"fromHeadImgUrl"`
	FromGender     int    `json:"fromGender"`
	FromGrade      int    `json:"fromGrade"`

	FromSchoolId   int    `json:"schoolId"`
	FromSchoolName string `json:"schoolName"`
	FromSchoolType int    `json:"schoolType"`

	SrcType int    `json:"srcType"`
	MsgDesc string `json:"msgDesc"`
	Status  int    `json:"status"` //0未处理 1同意已经是好友 2用户忽略
	Ts      int    `json:"ts"`
}

func (this *AddFriendMsgInfo) String() string {

	return fmt.Sprintf(`AddFriendMsgInfo{MsgId:%d, Uin:%d, FromUin:%d, FromNickName:%s, FromHeadImgUrl:%s, FromGender:%d, FromGrade:%d, FromSchoolId:%d, FromSchoolType:%d, FromSchoolName:%s, MsgDesc:%s, Status:%d, Ts:%d}`,
		this.MsgId, this.Uin, this.FromUin, this.FromNickName, this.FromHeadImgUrl, this.FromGender, this.FromGrade, this.FromSchoolId, this.FromSchoolType, this.FromSchoolName, this.MsgDesc, this.Status, this.Ts)
}

type GetAddFriendMsgRsp struct {
	Total int                 `json:"total"`
	Msgs  []*AddFriendMsgInfo `json:"msgs"`
}

func doGetAddFriendMsg(req *GetAddFriendMsgReq, r *http.Request) (rsp *GetAddFriendMsgRsp, err error) {

	log.Debugf("uin %d, GetAddFriendMsgReq %+v", req.Uin, req)

	total, msgs, err := GetAddFriendMsg(req.Uin, req.PageNum, req.PageSize, req.UpdateLastReadMsgId)

	if err != nil {
		log.Errorf("uin %d, GetAddFriendMsgRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetAddFriendMsgRsp{total, msgs}

	log.Debugf("uin %d, GetAddFriendMsgRsp succ, %+v", req.Uin, rsp)

	return
}

func GetAddFriendMsg(uin int64, pageNum, pageSize int, updateLastReadMsgId int) (total int, msgs []*AddFriendMsgInfo, err error) {

	msgs = make([]*AddFriendMsgInfo, 0)
	if uin == 0 {
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//只查询未处理的消息
	sql := fmt.Sprintf(`select count(msgId) from addFriendMsg where toUin = %d and status = 0`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	//只查询未处理的消息
	sql = fmt.Sprintf(`select msgId, fromUin, srcType, status, ts from addFriendMsg where toUin = %d and status = 0 order by msgId desc limit %d, %d `, uin, s, e)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	uins := make([]int64, 0)

	for rows.Next() {
		var msg AddFriendMsgInfo

		msg.Uin = uin
		rows.Scan(&msg.MsgId, &msg.FromUin, &msg.SrcType, &msg.Status, &msg.Ts)

		msgs = append(msgs, &msg)
		uins = append(uins, msg.FromUin)
	}

	res, err := st.BatchGetUserProfileInfo(uins)
	if err != nil {
		return
	}

	for i, msg := range msgs {

		if ui, ok := res[msg.FromUin]; ok {

			msg.FromNickName = ui.NickName
			msg.FromHeadImgUrl = ui.HeadImgUrl
			msg.FromGender = ui.Gender
			msg.FromGrade = ui.Grade
			msg.FromSchoolId = ui.SchoolId
			msg.FromSchoolName = ui.SchoolName
			msg.FromSchoolType = ui.SchoolType

			if msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_REGISTED || msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_NOT_REGISTED {

				msg.MsgDesc = "来自TA的通讯录"

			} else if msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL || msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GRADE || msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_BOY || msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GIRL {

				msg.MsgDesc = fmt.Sprintf("同校%s", st.GetGradeDescBySchool(msg.FromSchoolType, msg.FromGrade))

			} else if msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_TYPE_2DEGREE_FRIEND {

				msg.MsgDesc = "来自好友的好友"

			} else if msg.SrcType == constant.ENUM_RECOMMEND_FRIEND_SEARCH {

				msg.MsgDesc = "来自搜索"
			}

			msgs[i] = msg
		}
	}

	//if  updateLastReadMsgId > 0{
	{

		app, err1 := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_READ_ADDFRIEND_MSG_ID)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}

		keyStr := fmt.Sprintf("%d", uin)
		valStr, err1 := app.Get(keyStr)
		if err1 != nil {

			//如果KEY不存在 则认为lastMsgId为0
			if e, ok := err1.(*rest.APIError); ok {
				if e.Code == constant.E_REDIS_KEY_NO_EXIST {
					valStr = "0"
				} else {
					log.Error(err1.Error())
					return
				}
			} else {
				log.Error(err1.Error())
				return
			}
		}

		lastMsgId, err1 := strconv.Atoi(valStr)
		if err1 != nil {
			log.Error(err1.Error())
			lastMsgId = 0
		}

		curMsgId := 0
		if len(msgs) > 0 {
			curMsgId = int(msgs[0].MsgId)
		}

		if curMsgId <= lastMsgId {
			return
		}

		err1 = app.Set(keyStr, fmt.Sprintf("%d", curMsgId))
		if err1 != nil {
			log.Error(err1.Error())
		}
	}

	return
}
