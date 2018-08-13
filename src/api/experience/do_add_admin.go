package experience

import (
	"net/http"
)

type AddAdminReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int   `schema:"boardId"`
	LabelId  int   `schema:"labelId"`
	AdminUid int64 `schema:"adminUid"`
}

type AddAdminRsp struct {
}

func doAddAdmin(req *AddAdminReq, r *http.Request) (rsp *AddAdminRsp, err error) {

	log.Debugf("uin %d, AddAdminReq succ, %+v", req.Uin, rsp)

	err = AddAdmin(req.Uin, req.BoardId, req.LabelId, req.AdminUid)

	if err != nil {
		log.Errorf("uin %d, AddAdminReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddAdminRsp{}
	return
}
