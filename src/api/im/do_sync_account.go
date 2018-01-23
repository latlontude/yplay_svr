package im

import (
	"bytes"
	"common/constant"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	//"encoding/base64"
)

type SyncAccountReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Identifier string `schema:"identifier"`
	NickName   string `schema:"nickName"`
}

type ImSyncData struct {
	Identififer string `json:"Identifier"`
	NickName    string `json:"Nick"`
}

type ImSyncRsp struct {
	ActionStatus string `json:"ActionStatus"`
	ErrorCode    int    `json:"ErrorCode"`
	ErrorInfo    string `json:"ErrorInfo"`
}

type SyncAccountRsp struct {
}

func doSyncAccount(req *SyncAccountReq, r *http.Request) (rsp *SyncAccountRsp, err error) {

	log.Debugf("uin %d, SyncAccountReq %+v", req.Uin, req)

	err = SyncAccount(req.Uin, req.Identifier, req.NickName)
	if err != nil {
		log.Errorf("uin %d, SyncAccountRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &SyncAccountRsp{}

	log.Debugf("uin %d, SyncAccountRsp succ, %+v", req.Uin, rsp)

	return
}

func SyncAccount(uin int64, identifier string, nickName string) (err error) {

	if uin == 0 || len(identifier) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	dataSt := ImSyncData{identifier, nickName}

	log.Errorf("IMSyncAccountReq uin %d, req %+v", uin, dataSt)

	data, err := json.Marshal(&dataSt)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SYNC_ACCOUNT, err.Error())
		log.Errorf(err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/im_open_login_svc/account_import?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
		sig, constant.ENUM_IM_IDENTIFIER_ADMIN, constant.ENUM_IM_SDK_APPID, time.Now().Unix())

	rsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SYNC_ACCOUNT, err.Error())
		log.Errorf(err.Error())
		return
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SYNC_ACCOUNT, err.Error())
		log.Errorf(err.Error())
		return
	}

	var imSyncRsp ImSyncRsp

	err = json.Unmarshal(body, &imSyncRsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SYNC_ACCOUNT, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("ImSyncAccountRsp uin %d, rsp %+v", uin, imSyncRsp)

	if imSyncRsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SYNC_ACCOUNT, imSyncRsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
