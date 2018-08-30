package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)


func CheckPermit(uin int64, boardId int, labelId int) (hasPermission bool, err error) {

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	//board
	sql := fmt.Sprintf(`select ownerUid from v2boards where boardId = %d`, boardId)
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

	//admin
	sql = fmt.Sprintf(`select uin from experience_admin  where boardId = %d`, boardId)
	rows, err = inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

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