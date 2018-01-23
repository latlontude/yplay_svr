package cache

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

type QIconInfo struct {
	QId      int    `json:"qiconId"`
	QIconUrl string `json:"qiconUrl"`
	Ts       int    `json:"ts"`
}

func CacheQIcons() (err error) {

	QICONS = make(map[int]*QIconInfo)

	QICONSMAXTS = 0

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select qiconId, qiconUrl, ts from qicons`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var qid int
		var qiconUrl string
		var ts int

		rows.Scan(&qid, &qiconUrl, &ts)

		QICONS[qid] = &QIconInfo{qid, qiconUrl, ts}

		if ts > QICONSMAXTS {
			QICONSMAXTS = ts
		}
	}

	return
}
