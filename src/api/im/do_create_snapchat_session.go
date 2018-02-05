package im

import (
	"bytes"
	"common/constant"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//yplay创建群组请求
type CreateSnapChatSessonReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	User int64 `schema:"user"`
}

//yplay创建群组相应
type CreateSnapChatSessonRsp struct {
	SessionId string `json:"sessionId"`
}

func doCreateSnapChatSession(req *CreateSnapChatSessonReq, r *http.Request) (rsp *CreateSnapChatSessonRsp, err error) {

	log.Debugf("uin %d, CreateSnapChatSessonReq %+v", req.Uin, req)

	sessionId, err := CreateSnapChatSesson(req.Uin, req.User)
	if err != nil {
		log.Errorf("uin %d, CreateSnapChatSessonRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &CreateSnapChatSessonRsp{sessionId}

	log.Debugf("uin %d, CreateSnapChatSessonRsp succ, %+v", req.Uin, rsp)

	return
}

func MakeIMCreateSnapChatSessonReq(uin int64, user int64, groupName string) (req IMCreateGroupReq, err error) {

	req.Name = fmt.Sprintf("%s", groupName)
	//req.Name = ""
	req.Type = "Private"
	req.MemberList = make([]MemberInfo, 0)

	req.MemberList = append(req.MemberList, MemberInfo{fmt.Sprintf("%d", uin)})
	req.MemberList = append(req.MemberList, MemberInfo{fmt.Sprintf("%d", user)})

	return
}

func CreateSnapChatSesson(uin int64, user int64) (sessionId string, err error) {

	if uin == 0 || user == 0 || uin == user {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SNAPCHAT_SESSION)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_%d", uin, user)
	if uin > user {
		keyStr = fmt.Sprintf("%d_%d", user, uin)
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
		sessionId = strings.TrimSpace(valStr)
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	groupName := "SnapChatSession"

	req, err := MakeIMCreateSnapChatSessonReq(uin, user, groupName)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("uin %d, user %d, IMCreateSnapChatSessonReq %+v", uin, user, req)

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

	log.Errorf("uin %d, user %d, IMCreateSnapChatSessonRsp %+v", uin, user, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_CREATE_GROUP, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	sessionId = rsp.GroupId

	//设置redis失败认为是失败
	err = app.Set(keyStr, sessionId)
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
