package like

import (
	"common/constant"
	"common/rest"
	"database/sql"
	"fmt"
	"time"
)


/**
	点赞
 */
func insertV2Like(uin int64, likeId int, qid int, typ int ,inst *sql.DB ) (err error) {

	if inst == nil {
		return
	}

	sql := fmt.Sprintf(`insert into v2likes(id, qid, type, likeId, ownerUid, likeStatus, likeTs) 
		values(?, ?, ?, ?, ?, ?, ?)`)

	stmt, err := inst.Prepare(sql)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	status := 0 //0 默认
	_, err = stmt.Exec(0, qid, typ, likeId, uin, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("insert like By id")
	return
}



/**
	更新likeStatus
 */

func updateV2LikeById(id int, likeStatus int, inst *sql.DB )(err error) {

	sql := fmt.Sprintf(`update v2likes set likeStatus = %d where id = %d`,likeStatus,id)
	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
	}
	log.Debugf("update like By id")
	return
}


func updateV2Like(uin int64 , qid int , typ int , likeId int , likeStatus int,inst *sql.DB)(err error)  {
	sql := fmt.Sprintf(`update v2likes set likeStatus = %d where qid = %d and type = %d and likeId = %d and ownerUid = %d`,
		likeStatus,qid,typ,likeId,uin)

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
	}

	log.Debugf("update like ")

	return
}

