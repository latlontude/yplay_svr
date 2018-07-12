package experience

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelQidInExperienceReq struct {

	Uin         int64   `schema:"uin"`
	Token       string  `schema:"token"`
	Ver         int     `schema:"ver"`
	Qid         int     `schema:"qid"`
	LabelId     int     `schema:"labelId"`
	BoardId     int     `schema:"boardId"`

}

type DelQidInExperienceRsp struct {

}



func doDelQidInExperience(req *DelQidInExperienceReq, r *http.Request) (rsp *DelQidInExperienceRsp, err error) {

	log.Debugf("uin %d, DelQidInExperienceReq succ, %+v", req.Uin, rsp)

	err  = DelQidInExperience(req.Uin, req.BoardId, req.Qid, req.LabelId)

	if err != nil {
		log.Errorf("uin %d, DelQidInExperienceReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelQidInExperienceRsp{}
	return
}


func DelQidInExperience(uin int64, boardId,qid,labelId int) (err error){

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

	sql := fmt.Sprintf(`delete from experience_share where boardId = %d and labelId = %d and qid = %d`,boardId,labelId,qid)

	rows, err := inst.Query(sql)

	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	return
}


