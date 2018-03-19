package story

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type GetStoryVideoUploadSigReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetStoryVideoUploadSigRsp struct {
	Sig      string `json:"sig"`
	ExpireAt int    `json:"expireAt"`
}

func doGetStoryVideoUploadSig(req *GetStoryVideoUploadSigReq, r *http.Request) (rsp *GetStoryVideoUploadSigRsp, err error) {

	log.Debugf("uin %d, GetStoryVideoUploadSigReq %+v", req.Uin, req)

	sig, expireAt, err := GetStoryVideoUploadSig(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetStoryVideoUploadSigRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &GetStoryVideoUploadSigRsp{sig, expireAt}

	log.Debugf("uin %d, GetStoryVideoUploadSigRsp succ, %+v", req.Uin, rsp)

	return
}

func GetStoryVideoUploadSig(uin int64) (sig string, expireAt int, err error) {

	SECRETID := "AKIDrZ1EG40z72i7LKsUfaFfoiMmywffo4PV"
	SECRETKEY := "0lHZcQzJODm59xVnvoUyGK9j4A1cw0sV"

	now := int(time.Now().Unix())
	expireAt = now + 80*24*3600

	src := fmt.Sprintf(`secretId=%d&currentTimeStamp=%d&expireTime=%d&random=%d`, SECRETID, now, expireAt, rand.Intn(1000000))

	mac := hmac.New(sha1.New, []byte(SECRETKEY))
	mac.Write([]byte(src))
	hash := mac.Sum(nil)
	src2 := fmt.Sprintf("%s%s", string(hash), string(src))
	sig = base64.StdEncoding.EncodeToString([]byte(src2))

	return
}
