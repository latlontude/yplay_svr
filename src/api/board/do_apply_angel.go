package board

import (
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type ApplyAngelReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId int `schema:"boardId"`
}

type ApplyAngelRsp struct {
	Code int `json:"code"`
}

func doApplyAngel(req *ApplyAngelReq, r *http.Request) (rsp *ApplyAngelRsp, err error) {

	log.Debugf("apply angel req:%v,uin:%d", req, req.Uin)

	err = ApplyAngel(req.Uin, req.BoardId)
	code := 0
	rsp = &ApplyAngelRsp{code}
	return

}

func ApplyAngel(uin int64, boardId int) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select msgId, status,result , applyTs,dealTs from apply_angel 
where status = 0 and boardId = %d and  applyUin = %d`, boardId, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	var status int
	var msgId, applyTs,dealTs int64
	for rows.Next() {
		rows.Scan(&msgId,&status,&applyTs,&dealTs)
	}

	if msgId > 0 {
		//已经申请过了 正在处理中
	}

	//3.插入表
	stmt, err := inst.Prepare(`insert into apply_angel values(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()
	applyTs = time.Now().Unix()
	res, err := stmt.Exec(0, boardId, uin, 0, 0 ,applyTs, 0)        //id boardId , uin status result ts ts
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//获取新增数据id
	msgId, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	//4.发送通知
	go v2push.SendApplyAngelPush(uin,boardId, int(msgId))
	return
}
