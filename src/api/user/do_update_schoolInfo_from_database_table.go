package user

import (
	"bytes"
	"common/constant"
	"common/mydb"
	"common/mymgo"
	"common/rest"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

type UpdateSchoolReq struct {
	Uin      int64 `schema:"uin"`
	SchoolId int   `schema:"schoolId"`
}

type UpdateSchoolRsp struct {
}

func doUpdateSchool(req *UpdateSchoolReq, r *http.Request) (rsp *UpdateSchoolRsp, err error) {
	log.Debugf("uin %d, doUpdateSchool schoolId:%d", req.Uin, req.SchoolId)

	err = UpdateSchool(req.Uin, req.SchoolId)
	if err != nil {
		log.Error(err.Error())
		return
	}
	rsp = &UpdateSchoolRsp{}

	log.Debugf("uin %d, doUpdateSchool succ", req.Uin)

	return
}

func UpdateSchool(uin int64, schoolId int) (err error) {
	log.Debugf("start UpdateSchool uin:%d schoolId:%d", uin, schoolId)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select schoolType, schoolName, country, province, city, latitude, longitude, status from schools where schoolId = %d`, schoolId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false

	var schoolType int
	var schoolName string
	var country string
	var province string
	var city string
	var latitude float64
	var longitude float64
	var status int
	ts := time.Now().Unix()

	for rows.Next() {
		rows.Scan(&schoolType, &schoolName, &country, &province, &city, &latitude, &longitude, &status)
		find = true
	}

	if !find {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid schoolId")
		log.Errorf(err.Error())
		return
	}

	if schoolType > constant.ENUM_SCHOOL_TYPE_UNIVERSITY || schoolType < 0 || len(schoolName) == 0 || len(country) == 0 || len(province) == 0 || len(city) == 0 || latitude < 1.0 || longitude < 1.0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "school information incomplete!")
		log.Error(err.Error())
		return
	}

	//将学校信息存储到mgo中以便于根据位置查找附近的学校
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

	mgoSchoolInfoMap := make(map[string]interface{})
	mgoLocalINfoMap := make(map[string]interface{})

	mgoLocalINfoMap["type"] = "Point"
	mgoLocalINfoMap["coordinates"] = []float64{longitude, latitude}

	mgoSchoolInfoMap["_id"] = schoolId
	mgoSchoolInfoMap["type"] = schoolType
	mgoSchoolInfoMap["name"] = schoolName
	mgoSchoolInfoMap["province"] = province
	mgoSchoolInfoMap["city"] = city
	mgoSchoolInfoMap["loc"] = mgoLocalINfoMap

	log.Debugf("mgoSchoolInfoMap %+v", mgoSchoolInfoMap)

	err = c.Insert(&mgoSchoolInfoMap)
	if err != nil {
		log.Debugf(err.Error())
		log.Debugf("update doc")
		err = c.Update(&map[string]interface{}{"_id": schoolId}, &mgoSchoolInfoMap)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	//将学校信息存储到es中，根据学校名搜索时使用

	esSchoolInfoMap := make(map[string]interface{})
	esSchoolInfoMap["schoolId"] = schoolId
	esSchoolInfoMap["schoolType"] = schoolType
	esSchoolInfoMap["school"] = schoolName
	esSchoolInfoMap["country"] = country
	esSchoolInfoMap["province"] = province
	esSchoolInfoMap["city"] = city
	esSchoolInfoMap["latitude"] = latitude
	esSchoolInfoMap["longitude"] = longitude
	esSchoolInfoMap["status"] = status
	esSchoolInfoMap["ts"] = ts

	d, err := json.Marshal(&esSchoolInfoMap)
	if err != nil {
		log.Errorf("json marshal error", err.Error())
		return
	}

	log.Debugf("len d :%d d:%+v", len(d), string(d))
	client := &http.Client{}
	url := fmt.Sprintf("http://122.152.206.97:9200/yplay/schools/%d", schoolId)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(d))
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	client.Do(req)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(d)))
	req.Header.Add("X-Content-Length", strconv.Itoa(len(d)))
	dump, err := httputil.DumpRequest(req, true)

	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Debugf("dump:%s", string(dump))

	sql = fmt.Sprintf("update schools set ts = %d where schoolId = %d ", ts, schoolId)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("end UpdateSchool")
	return
}
