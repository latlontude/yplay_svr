package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
	"time"
)

type AddAdminReq struct {
	Uin         int64   `schema:"uin"`
	Token       string  `schema:"token"`
	Ver         int     `schema:"ver"`

	BoardId     int     `schema:"boardId"`
	LabelId     int     `schema:"labelId"`
	AdminUid    int64   `schema:"adminUid"`

}

type AddAdminRsp struct {

}



func doAddAdmin(req *AddAdminReq, r *http.Request) (rsp *AddAdminRsp, err error) {

	log.Debugf("uin %d, AddAdminReq succ, %+v", req.Uin, rsp)

	err  = AddAdmin(req.Uin,req.BoardId,req.LabelId,req.AdminUid)

	if err != nil {
		log.Errorf("uin %d, AddAdminReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddAdminRsp{}
	return
}


func AddAdmin(uin int64, boardId int ,labelId  int , adminUid int64 ) (err error){
	if uin < 0  {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into experience_admin(id, boardId, labelId,uin,ts) values(?, ?, ?, ?,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	_ , err = stmt.Exec(0, boardId, labelId,adminUid , ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}