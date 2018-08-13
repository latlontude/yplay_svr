package experience

import (
	"api/elastSearch"
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type DelQidInExperienceReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	Qid      int    `schema:"qid"`
	AnswerId int    `schema:"answerId"`
	LabelId  int    `schema:"labelId"`
	BoardId  int    `schema:"boardId"`
}

type DelQidInExperienceRsp struct {
}

func doDelAnswerIdFromExp(req *DelQidInExperienceReq, r *http.Request) (rsp *DelQidInExperienceRsp, err error) {

	log.Debugf("uin %d, DelQidInExperienceReq succ,  req :%+v", req.Uin, req)

	err = DelAnswerIdFromExp(req.Uin, req.BoardId, req.Qid, req.AnswerId, req.LabelId)

	if err != nil {
		log.Errorf("uin %d, DelQidInExperienceReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &DelQidInExperienceRsp{}
	return
}

func DelAnswerIdFromExp(uin int64, boardId, qid, answerId, labelId int) (err error) {

	//校验权限
	hasPermission, err := CheckPermit(uin, boardId, labelId)

	if !hasPermission {
		err = rest.NewAPIError(constant.E_DB_QUERY, "del answer has not  permit")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`delete from experience_share where boardId = %d and labelId = %d and answerId = %d`, boardId, labelId, answerId)

	rows, err := inst.Query(sql)

	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	err1 := elastSearch.DelAnswerFromEs(answerId)
	if err1 != nil {
		log.Debugf("es delete label error")
	}

	arrayUin := []int64{102772, 102773, 102774, 103307, 103122, 103126, 103096, 103004, 103032, 101749}

	for _, v := range arrayUin {
		if uin == v {
			go v2push.SendDelAnswerIdInExpPush(uin, qid, answerId, labelId)

		}
	}

	return
}

//删除回答  同时从经验弹中删除
func DelAnswerFromExpByAnswerId(uin int64, answerId int) (err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select boardId,labelId,qid from experience_share where answerId = %d`, answerId)

	rows, err := inst.Query(sql)

	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var boardId, labelId, qid int
		rows.Scan(&boardId, &labelId, &qid)
		DelAnswerIdFromExp(uin, boardId, qid, answerId, labelId)
	}
	return
}
