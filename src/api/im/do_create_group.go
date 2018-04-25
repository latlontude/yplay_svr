package im

import (
	"bytes"
	"common/constant"
	"common/mydb"
	//"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"strings"
	"svr/st"
	"time"
)

//IM后台的创建群组请求
type IMCreateGroupReq struct {
	Name       string       `json:"Name"`
	Type       string       `json:"Type"`
	MemberList []MemberInfo `json:"MemberList"`
}

type MemberInfo struct {
	MemberAccount string `json:"Member_Account"`
}

//IM后台的创建群组相应包
type IMCreateGroupRsp struct {
	ActionStatus string `json:"ActionStatus"`
	ErrorInfo    string `json:"ErrorInfo"`
	ErrorCode    int    `json:"ErrorCode"`
	GroupId      string `json:"GroupId"`
}

//yplay创建群组请求
type CreateGroupReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	VoteToUin    int64 `schema:"voteToUin"`
	VoteRecordId int64 `schema:"voteRecordId"`
}

//yplay创建群组相应
type CreateGroupRsp struct {
	GroupId string `json:"groupId"`
}

func doCreateGroup(req *CreateGroupReq, r *http.Request) (rsp *CreateGroupRsp, err error) {

	log.Debugf("uin %d, CreateGroupReq %+v", req.Uin, req)

	vinfo, err := st.GetVoteRecordInfo(req.VoteRecordId)
	if err != nil {
		log.Errorf("uin %d, CreateGroupRsp error, %s", req.Uin, err.Error())
		return
	}

	groupId, err := CreateGroup(req.Uin, req.VoteToUin, req.VoteRecordId, vinfo.QText)
	if err != nil {
		log.Errorf("uin %d, CreateGroupRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &CreateGroupRsp{groupId}

	log.Debugf("uin %d, CreateGroupRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMCreateGroupReq(uin int64, voteToUin int64, groupName string) (req IMCreateGroupReq, err error) {

	req.Name = fmt.Sprintf("%s", groupName)
	//req.Name = ""
	req.Type = "Private"
	req.MemberList = make([]MemberInfo, 0)

	req.MemberList = append(req.MemberList, MemberInfo{fmt.Sprintf("%d", uin)})
	req.MemberList = append(req.MemberList, MemberInfo{fmt.Sprintf("%d", voteToUin)})

	return
}

func CreateGroup(uin int64, voteToUin int64, voteRecordId int64, groupName string) (groupId string, err error) {

	if uin == 0 || voteToUin == 0 || len(groupName) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	if voteToUin == 0 {
		return
	}

	if uin == voteToUin {
		return
	}
	/*
		app, err := myredis.GetApp(constant.ENUM_REDIS_APP_VOTECHAT_SESSION)
		if err != nil {
			log.Error(err.Error())
			return
		}

		keyStr := fmt.Sprintf("%d_%d", uin, voteToUin)
		if uin > voteToUin {
			keyStr = fmt.Sprintf("%d_%d", voteToUin, uin)
		}

		valStr, err := app.Get(keyStr)

		if err != nil {

			//如果KEY不存在
			e, ok := err.(*rest.APIError)

			if !ok {
				log.Error(err.Error())
				return
			}

			if e.Code != constant.E_REDIS_KEY_NO_EXIST {
				log.Error(err.Error())
				return
			}

		} else {

			groupId = strings.TrimSpace(valStr)

			if len(groupId) > 0 {
				log.Errorf("uin %d, voteToUin %d, CreateGroup, return groupId from redis %s", uin, voteToUin, groupId)
				return
			}
		}
	*/
	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	newName := "噗噗"

	req, err := MakeIMCreateGroupReq(uin, voteToUin, newName)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("uin %d, voteToUin %d, voteRecordId %d, IMCreateGroupReq %+v", uin, voteToUin, voteRecordId, req)

	data, err := json.Marshal(&req)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_CREATE_GROUP, err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/group_open_http_svc/create_group?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
		sig, constant.ENUM_IM_IDENTIFIER_ADMIN, constant.ENUM_IM_SDK_APPID, time.Now().Unix())

	hrsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_CREATE_GROUP, err.Error())
		log.Errorf(err.Error())
		return
	}

	body, err := ioutil.ReadAll(hrsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_CREATE_GROUP, err.Error())
		log.Errorf(err.Error())
		return
	}

	var rsp IMCreateGroupRsp
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_CREATE_GROUP, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("uin %d, voteToUin %d, voteRecordId %d, IMCreateGroupRsp %+v", uin, voteToUin, voteRecordId, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_CREATE_GROUP, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	groupId = rsp.GroupId
	/*
		//设置redis失败不认为是失败
		err1 := app.Set(keyStr, groupId)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}
	*/
	if uin > voteToUin {
		go storeVoteSessionId(voteToUin, uin, groupId)
	} else {
		go storeVoteSessionId(uin, voteToUin, groupId)
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	q := fmt.Sprintf(`update voteRecords set imSessionId = "%s" where id = %d `, groupId, voteRecordId)
	_, err = inst.Exec(q)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func storeVoteSessionId(uin, uid int64, sessionId string) (err error) {
	log.Debugf("start storeVoteSessionId uin:%d uid:%d sessionId:%s", uin, uid, sessionId)

	if uin == uid {
		log.Errorf("err uin equel uid")
		return
	}

	ts := time.Now().Unix()
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into voteSession values(?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, uin, uid, sessionId, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("end storeVoteSessionId uin:%d uid:%d sessionId:%s", uin, uid, sessionId)
	return
}
