package ddactivity

import (
	"net/http"
)

type GetCallTypeInfoReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetCallTypeInfoRsp struct {
	Singer    SingerInfo     `json:"singer"`  //爱豆个人信息
	CallInfos []CallTypeInfo `json:callInfos` // 各种打call类型完成情况信息数组
}

func doGetCallTypeInfo(req *GetCallTypeInfoReq, r *http.Request) (rsp *GetCallTypeInfoRsp, err error) {
	log.Debugf("start doGetCallTypeInfo uin:%d ", req.Uin)

	singer, err := getSingerInfo(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	callInfos, err := getCallTypeInfos(req.Uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetCallTypeInfoRsp{singer, callInfos}
	log.Debugf("end doGetCallTypeInfo  rsp:%+v", rsp)
	return
}
