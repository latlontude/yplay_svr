package notify

import (
	"api/feed"
	"api/sns"
	"api/submit"
	"net/http"
)

type GetNewNotifyStatReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetNewNotifyStatRsp struct {
	NewFeedCnt                int `json:"newFeedCnt"`
	NewAddFriendMsgCnt        int `json:"newAddFriendMsgCnt"`
	NewOnlineSubmitQustionCnt int `json:"newOnlineSubmitQustionCnt"`
}

func doGetNewNotifyStat(req *GetNewNotifyStatReq, r *http.Request) (rsp *GetNewNotifyStatRsp, err error) {

	log.Debugf("uin %d, GetNewNotifyStatReq %+v", req.Uin, req)

	feedCnt, newAddFriendMsgCnt, newOnlineSubmitQustionCnt, err := GetNewNotifyStat(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetNewNotifyStatRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetNewNotifyStatRsp{feedCnt, newAddFriendMsgCnt, newOnlineSubmitQustionCnt}

	log.Debugf("uin %d, GetNewNotifyStatRsp succ, %+v", req.Uin, rsp)

	return
}

func GetNewNotifyStat(uin int64) (newFeedCnt, newAddFriendMsgCnt, newOnlineSubmitQustionCnt int, err error) {

	newFeedCnt, err = feed.GetNewFeedsCnt(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	newAddFriendMsgCnt, err = sns.GetAddFriendNewMsgCnt(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	newOnlineSubmitQustionCnt, err = submit.SubmitGetNewOnlineCnt(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
