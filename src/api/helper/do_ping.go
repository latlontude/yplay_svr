package helper

import (
	"net/http"
	//"time"
)

type PingReq struct {
}

type PingRsp struct {
}

func doPing(req *PingReq, r *http.Request) (rsp *PingRsp, err error) {

	log.Debugf("PingReq header %+v", r.Header)

	rsp = &PingRsp{}

	//time.Sleep(5 * time.Second)

	log.Debugf("PingRsp")

	return
}
