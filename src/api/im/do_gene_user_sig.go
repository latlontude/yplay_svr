package im

import (
	"common/constant"
	"common/rest"
	"encoding/base64"
	"fmt"
	"net/http"
	//"encoding/json"
	"time"
)

type GeneUserSigReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Identifier string `schema:"identifier"`
}

type GeneUserSigRsp struct {
	Sig      string `json:"sig"`
	ExpireAt int    `json:"expireAt"`
}

func doGeneUserSig(req *GeneUserSigReq, r *http.Request) (rsp *GeneUserSigRsp, err error) {

	log.Debugf("uin %d, GeneUserSigReq %+v", req.Uin, req)

	sig, expireAt, err := GeneUserSig(req.Uin, req.Identifier)
	if err != nil {
		log.Errorf("uin %d, GeneUserSigRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GeneUserSigRsp{sig, expireAt}

	log.Debugf("uin %d, GeneUserSigRsp succ, %+v", req.Uin, rsp)

	return
}

func GeneUserSig(uin int64, identifier string) (sig string, expireAt int, err error) {

	expireAt = int(time.Now().Unix()) + 170*86400

	if uin == 0 || len(identifier) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	var conf TLSSignConf

	priKey, err := base64.StdEncoding.DecodeString(Base64PriKeyString)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Errorf(err.Error())
		return
	}

	conf.PriKey = string(priKey)

	conf.SDKAppId = constant.ENUM_IM_SDK_APPID
	conf.Identifier = fmt.Sprintf("%s", identifier)

	sig, err = conf.Sign()
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_GENE_SIG, err.Error())
		log.Errorf(err.Error())
	}

	return
}

func GetAdminSig() (sig string, err error) {

	ts := int(time.Now().Unix())

	if len(IM_SIG_ADMIN) > 0 && (ts-IM_SIG_ADMIN_GENE_TS) <= 170*86400 {
		sig = IM_SIG_ADMIN
		return
	}

	var conf TLSSignConf

	priKey, err := base64.StdEncoding.DecodeString(Base64PriKeyString)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Errorf(err.Error())
		return
	}

	conf.PriKey = string(priKey)

	conf.SDKAppId = constant.ENUM_IM_SDK_APPID
	conf.Identifier = fmt.Sprintf("%s", constant.ENUM_IM_IDENTIFIER_ADMIN)

	sig, err = conf.Sign()
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_GENE_SIG, err.Error())
		log.Errorf(err.Error())
	}

	IM_SIG_ADMIN = sig
	IM_SIG_ADMIN_GENE_TS = ts

	return
}
