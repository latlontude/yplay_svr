package cache

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

func CachePhones() (err error) {

	PHONE2UIN = make(map[string]int64)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, phone from profiles where nickName != "" `)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var phone string
		var uin int64

		rows.Scan(&uin, &phone)

		PHONE2UIN[phone] = uin
	}

	return
}

func UpdatePhone2Uin(phone string, uin int64) {
	if uin == 0 || len(phone) == 0 {
		return
	}

	PHONE2UIN[phone] = uin

	return
}
