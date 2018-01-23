package account

import (
	"common/constant"
	"common/mydb"
	"common/mymgo"
	"common/rest"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"svr/st"
)

type GetNearestSchoolsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	SchoolType int     `schema:"schoolType"`
	Latitude   float64 `schema:"latitude"`
	Longitude  float64 `schema:"longitude"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetNearestSchoolsRsp struct {
	Schools []*st.SchoolInfo2 `json:"schools"`
}

type GeoJson struct {
	Type        string    `json:"string"`
	Coordinates []float64 `json:"coordinates"`
}

type MgoSchoolInfo struct {
	SchoolId   int     `bson:"_id,omitempty" json:"schoolId"`
	SchoolName string  `bson:"name" json:"schoolName"`
	SchoolType int     `bson:"type" json:"schoolType"`
	Province   string  `bson:"province" json:"province"`
	City       string  `bson:"city" json:"city"`
	Loc        GeoJson `bson:"loc"  json:"loc"`
}

func doGetNearestSchools(req *GetNearestSchoolsReq, r *http.Request) (rsp *GetNearestSchoolsRsp, err error) {

	log.Debugf("uin %d, GetNearestSchoolsReq %+v", req.Uin, req)

	schools, err := GetNearestSchools2(req.Uin, req.SchoolType, req.Latitude, req.Longitude, req.PageNum, req.PageSize)
	if err != nil {
		log.Errorf("uin %d, GetNearestSchoolsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetNearestSchoolsRsp{schools}

	log.Debugf("uin %d, GetNearestSchoolsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetNearestSchools(uin int64, schoolType int, latitude, longitude float64, pageNum, pageSize int) (schools []*st.SchoolInfo2, err error) {

	schools = make([]*st.SchoolInfo2, 0)

	if uin == 0 || schoolType < constant.ENUM_SHOOL_TYPE_JUNIOR || schoolType > constant.ENUM_SCHOOL_TYPE_UNIVERSITY {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageSize

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select schoolId, schoolType, schoolName, country, province, city, latitude, longitude, status, ts from schools where schoolType = %d limit %d, %d`, schoolType, s, e)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	schoolIds := make(map[int]int, 0)

	for rows.Next() {
		var info st.SchoolInfo2
		rows.Scan(&info.SchoolId, &info.SchoolType, &info.SchoolName, &info.Country, &info.Province, &info.City, &info.Latitude, &info.Longitude, &info.Status, &info.Ts)

		schools = append(schools, &info)

		schoolIds[info.SchoolId] = 1
	}

	if len(schools) == 0 {
		return
	}

	schoolsStr := ""
	for sid, _ := range schoolIds {
		schoolsStr += fmt.Sprintf("%d,", sid)
	}

	schoolsStr = schoolsStr[:len(schoolsStr)-1]

	sql = fmt.Sprintf(`select schoolId, count(uin) as cnt from profiles where schoolId in (%s) group by schoolId`, schoolsStr)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	members := make(map[int]int)
	for rows.Next() {
		var schoolId, cnt int
		rows.Scan(&schoolId, &cnt)

		members[schoolId] = cnt
	}

	for i, si := range schools {

		if v, ok := members[si.SchoolId]; ok {
			si.MemberCnt = v

			schools[i] = si
		}
	}

	return
}

func GetNearestSchools2(uin int64, schoolType int, latitude, longitude float64, pageNum, pageSize int) (schools []*st.SchoolInfo2, err error) {

	if uin == 0 || latitude > 90 || latitude < -90 || longitude > 180 || longitude < -180 {
		return
	}

	schools = make([]*st.SchoolInfo2, 0)

	//都是Clone出来的, 要及时Close
	session := mymgo.GetSession(constant.ENUM_MGO_INST_YPLAY)
	if session == nil {
		err = rest.NewAPIError(constant.E_MGO_INST_NIL, "mgo clone session nil")
		log.Error(err.Error())
		return
	}
	defer session.Close()

	db := session.DB("yplay")
	if db == nil {
		err = rest.NewAPIError(constant.E_MGO_DB_NIL, "mgo database nil")
		log.Error(err.Error())
		return
	}

	c := db.C("schools")
	if c == nil {
		err = rest.NewAPIError(constant.E_MGO_COLLECTION_NIL, "mgo collection nil")
		log.Error(err.Error())
		return
	}

	mgoSchools := make([]MgoSchoolInfo, 0)

	err = c.Find(bson.M{"loc": bson.M{"$near": bson.M{"type": "Point", "coordinates": []float64{longitude, latitude}}}, "type": schoolType}).Limit(100).All(&mgoSchools)
	if err != nil {
		err = rest.NewAPIError(constant.E_MGO_QUERY, err.Error())
		log.Error(err.Error())
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := pageNum * pageSize

	if s >= len(mgoSchools) {
		return
	}

	if e > len(mgoSchools) {
		e = len(mgoSchools)
	}

	retSchools := mgoSchools[s:e]

	if len(retSchools) == 0 {
		return
	}

	schoolsM := make(map[int]int)
	schoolsA := make([]int, 0)

	for _, msi := range retSchools {

		var si st.SchoolInfo2

		si.SchoolId = msi.SchoolId
		si.SchoolType = msi.SchoolType
		si.SchoolName = msi.SchoolName
		si.Province = msi.Province
		si.City = msi.City

		si.Latitude = msi.Loc.Coordinates[1]
		si.Longitude = msi.Loc.Coordinates[0]

		schools = append(schools, &si)

		if _, ok := schoolsM[si.SchoolId]; ok {
			continue
		}

		schoolsA = append(schoolsA, si.SchoolId)

		schoolsM[si.SchoolId] = 1
	}

	//现在版本先不显示成员数，后续待看
	/*
		res, err1 := GetSchoolsMemerCnt(schoolsA)
		if err1 != nil {
			log.Error(err1.Error())
			return
		}

		for i, si := range schools {

			if v, ok := res[si.SchoolId]; ok {
				si.MemberCnt = v
				schools[i] = si
			}
		}
	*/

	return
}

func GetSchoolsMemerCnt(schoolIds []int) (res map[int]int, err error) {

	res = make(map[int]int)

	if len(schoolIds) == 0 {
		return
	}

	schoolsStr := ""
	for sid, _ := range schoolIds {
		schoolsStr += fmt.Sprintf("%d,", sid)
	}
	schoolsStr = schoolsStr[:len(schoolsStr)-1]

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select schoolId, count(uin) as cnt from profiles where schoolId in (%s) group by schoolId`, schoolsStr)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var schoolId, cnt int
		rows.Scan(&schoolId, &cnt)

		res[schoolId] = cnt
	}

	return
}
