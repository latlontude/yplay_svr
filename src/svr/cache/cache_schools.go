package cache

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func CacheSchools() (err error) {

	SCHOOLS = make(map[int]*st.SchoolInfo)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts from schools`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info st.SchoolInfo
		rows.Scan(&info.SchoolId, &info.SchoolType, &info.SchoolName, &info.Country, &info.Province, &info.City, &info.Latitude, &info.Longitude, &info.Status, &info.Ts)

		SCHOOLS[info.SchoolId] = &info
	}

	return
}

func AddCacheSchool(schoolInfo st.SchoolInfo) {

	SCHOOLS[schoolInfo.SchoolId] = &schoolInfo
	return
}
