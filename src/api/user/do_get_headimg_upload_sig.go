package user

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type GetHeadImgUploadSigReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetHeadImgUploadSigRsp struct {
	Sig      string `json:"sig"`
	ExpireAt int    `json:"expireAt"`
}

func doGetHeadImgUploadSig(req *GetHeadImgUploadSigReq, r *http.Request) (rsp *GetHeadImgUploadSigRsp, err error) {

	log.Debugf("uin %d, GetHeadImgUploadSigReq %+v", req.Uin, req)

	sig, expireAt, err := GetHeadImgUploadSig(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetHeadImgUploadSigRsp error %s", req.Uin, err.Error())
		return
	}

	rsp = &GetHeadImgUploadSigRsp{sig, expireAt}

	log.Debugf("uin %d, GetHeadImgUploadSigRsp succ, %+v", req.Uin, rsp)

	return
}

func GetHeadImgUploadSig(uin int64) (sig string, expireAt int, err error) {

	APPID := 1253229355
	BUCKET := "yplay"
	SECRETID := "AKIDrZ1EG40z72i7LKsUfaFfoiMmywffo4PV"
	SECRETKEY := "0lHZcQzJODm59xVnvoUyGK9j4A1cw0sV"

	now := int(time.Now().Unix())
	expireAt = now + 90*24*3600

	src := fmt.Sprintf(`a=%d&b=%s&k=%s&e=%d&t=%d&r=%d&u=0&f=`, APPID, BUCKET, SECRETID, expireAt, now, rand.Intn(100))

	mac := hmac.New(sha1.New, []byte(SECRETKEY))
	mac.Write([]byte(src))
	hash := mac.Sum(nil)
	src2 := fmt.Sprintf("%s%s", string(hash), string(src))
	sig = base64.StdEncoding.EncodeToString([]byte(src2))

	return
}
