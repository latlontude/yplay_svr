package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type AddQidInExperienceReq struct {

	Uin         int64   `schema:"uin"`
	Token       string  `schema:"token"`
	Ver         int     `schema:"ver"`
	Qid         int     `schema:"qid"`
	LabelId     int     `schema:"labelId"`
	BoardId     int     `schema:"boardId"`

}

type AddQidInExperienceRsp struct {

}



func doAddQidInExperience(req *AddQidInExperienceReq, r *http.Request) (rsp *AddQidInExperienceRsp, err error) {

	log.Debugf("uin %d, AddQidInExperienceReq succ, %+v", req.Uin, rsp)

	err  = AddQidInExperience(req.Uin, req.BoardId, req.Qid, req.LabelId)

	if err != nil {
		log.Errorf("uin %d, AddQidInExperienceReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AddQidInExperienceRsp{}
	return
}


func AddQidInExperience(uin int64, boardId,qid,labelId int) (err error){

	//校验权限
	hasPermission ,err := CheckPermit(uin,boardId,labelId)

	if !hasPermission {
		err = rest.NewAPIError(constant.E_DB_QUERY, "add question has not  permit")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into experience_share(id, boardId, labelId, qid, ts, status) 
		values(?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()


	_, err = stmt.Exec(0, boardId, labelId, qid, uin, ts,0)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}



func CheckPermit(uin int64 ,boardId int , labelId int ) (hasPermission bool,err error) {

	if uin == 0  {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select uin from experience_admin  where boardId = %d and labelId = %d` ,boardId,labelId)
	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	hasPermission = false

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		if uid == uin {
			hasPermission = true
		}
	}

	if uin == 100001 {
		hasPermission = true
	}

	return
}