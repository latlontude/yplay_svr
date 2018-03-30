package ddactivity

import (
	"common/rest"
	"net/http"
)

type LoadPageReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

func doLoadPage(req *LoadPageReq, r *http.Request) (rsp *rest.DownloadInfo, err error) {

	log.Debugf("uin %d, LoadPage %+v", req.Uin, req)

	rsp = &rest.DownloadInfo{req.Uin, req.Token, req.Ver, "index.html"}

	log.Debugf("uin %d, LoadPageRsp succ, %+v", req.Uin, rsp)

	return
}
