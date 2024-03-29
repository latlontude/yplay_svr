package board

import (
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type CheckApplyReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	MsgId   int    `schema:"msgId"`
	BoardId int    `schema:"boardId"`
	Result  int    `schema:"result"` //审核结果 1同意 2不同意
}

type CheckApplyRsp struct {
	Code int `json:"code"`
}

func doCheckApply(req *CheckApplyReq, r *http.Request) (rsp *CheckApplyRsp, err error) {

	log.Debugf("check apply req:%v,uin:%d", req, req.Uin)
	err = CheckApply(req.Uin, req.BoardId, req.MsgId, req.Result)
	code := 0
	rsp = &CheckApplyRsp{code}
	return

}

func CheckApply(uin int64, boardId int, msgId int, result int) (err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select applyUin,status,result ,applyTs,dealTs from apply_angel 
where status = 0 and msgId = %d and boardId = %d `, msgId,boardId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	var status, originalResult int
	var applyUin, applyTs, dealTs int64
	for rows.Next() {
		rows.Scan(&applyUin, &status, &originalResult, &applyTs, &dealTs)
	}

	if originalResult > 0 {
		//已经处理过了
		err = rest.NewAPIError(constant.E_CHECK_APPLY_REPEAT, "repeat deal")
		return
	}

	//1.更改审核结果 2.更改v2board 3.添加到admin
	sqlArr := make([]string, 0)
	sqlArr = append(sqlArr, fmt.Sprintf(`update apply_angel set result=%d,dealTs=%d where msgId = %d`, result, dealTs, msgId))


	//是否有墙主 如果有墙主 B则变成小天使
	//同意变更墙主

	var boardOwnerUin int64

	if result == 1 {
		sql = fmt.Sprintf(`select ownerUid from v2boards where boardId = %d`,boardId)
		rows, err = inst.Query(sql)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
			return
		}

		for rows.Next() {
			rows.Scan(&boardOwnerUin)
		}
		//还没有墙主
		if boardOwnerUin == 0 {
			sqlArr = append(sqlArr, fmt.Sprintf(`update v2boards  set ownerUid = %d where boardId = %d`, applyUin, boardId))
		}
	}
	err = mydb.Exec(inst, sqlArr)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//同意审核之后 才加入admin表
	if result == 1{
		//加入admin表
		boardType := 0
		if boardOwnerUin == 0 {
			boardType = 1
			result = 3      //3.主天使  1.小天使 2.拒绝
		}
		err = AddAngelInAdmin(100001, boardId, 0, applyUin,boardType)
		if err != nil {
			log.Errorf("add angel err , uin:%d err:%+v", uin, err.Error())
			return
		}
	}

	//4.发送通知
	go v2push.SendCheckApplyPush(uin, applyUin, boardId, result)
	return
}
