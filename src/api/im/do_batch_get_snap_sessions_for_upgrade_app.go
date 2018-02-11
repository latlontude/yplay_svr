package im

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

//yplay创建群组请求
type BatchGetSnapSessonsForUpgradeAppReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Users string `schema:"users"`
}

type SnapChatSessionInfo struct {
	Uin1      int64  `json:"uin1"`
	Uin2      int64  `json:"uin2"`
	SessionId string `json:"sessionId"`
}

//yplay创建群组相应
type BatchGetSnapSessonsForUpgradeAppRsp struct {
	Sessions []*SnapChatSessionInfo `json:"sessions"`
}

func doBatchGetSnapSessionsForUpgradeApp(req *BatchGetSnapSessonsForUpgradeAppReq, r *http.Request) (rsp *BatchGetSnapSessonsForUpgradeAppRsp, err error) {

	log.Debugf("uin %d, BatchGetSnapSessonsFroUpgradeAppReq %+v", req.Uin, req)

	ss, err := BatchGetSnapSessonsForUpgradeApp(req.Uin, req.Users)
	if err != nil {
		log.Errorf("uin %d, CreateSnapChatSessonRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &BatchGetSnapSessonsForUpgradeAppRsp{ss}

	log.Debugf("uin %d, BatchGetSnapSessonsForUpgradeAppRsp succ", req.Uin)

	return
}

func BatchGetSnapSessonsForUpgradeApp(uin int64, users string) (ss []*SnapChatSessionInfo, err error) {

	if uin == 0 || len(users) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	var friends []int64

	err = json.Unmarshal([]byte(users), &friends)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, err.Error())
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SNAPCHAT_SESSION)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keys := make([]string, 0)

	for _, user := range friends {
		keyStr := fmt.Sprintf("%d_%d", uin, user)
		if uin > user {
			keyStr = fmt.Sprintf("%d_%d", user, uin)
		}

		keys = append(keys, keyStr)
	}

	if len(keys) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	vals, err := app.MGet(keys)

	if err != nil {
		log.Error(err.Error())
		return
	}

	for k, v := range vals {
		a := strings.Split(k, "_")
		if len(a) != 2 {
			continue
		}

		uin1, _ := strconv.Atoi(a[0])
		uin2, _ := strconv.Atoi(a[1])

		if int64(uin2) == uin {
			uin2 = uin1
			uin1 = int(uin)
		}

		ss = append(ss, &SnapChatSessionInfo{int64(uin1), int64(uin2), v})
	}

	return
}
