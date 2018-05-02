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

type AddSchoolReq struct {
	Uin        int64   `schema:"uin"`
	SchoolType int     `schema:"schoolType"`
	SchoolName string  `schema:"schoolName"`
	Country    string  `schema:"Country"`
	Province   string  `schema:"province"`
	City       string  `schema:"city"`
	Latitude   float64 `schema:"latitude"`
	Longitude  float64 `schema:"longitude"`
}

type AddSchoolRsp struct {
}

func doAddSchool(req *AddSchoolReq, r *http.Request) (rsp *AddSchoolRsp, err error) {

	log.Debugf("uin %d, doAddSchool %+v", req.Uin, req)
	err = AddSchool(req.Uin, req.SchoolType, req.SchoolName, req.Country, req.Province, req.City, req.Latitude, req.Longitude)
	if err != nil {
		log.Error(err.Error())
		return
	}

	rsp = &AddSchoolRsp{}

	log.Debugf("uin %d, doAddSchool succ, %+v", req.Uin, rsp)

	return
}

func AddSchool(uin int64, schoolType int, schoolName, country, province, city string, latitude, longitude float64) (err error) {
	log.Debugf("start AddSchool uin:%d", uin)

	if uin == 0 || schoolType > constant.ENUM_SCHOOL_TYPE_UNIVERSITY || schoolType < 0 || len(schoolName) == 0 || len(country) == 0 || len(province) == 0 || len(city) == 0 || latitude < 1.0 || longitude < 1.0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	// 将学校信息存储到schools数据库表中
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`insert into schools values(?, ?, ?, ?, ?, ?, ?, ?, ? ,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err)
		return
	}
	defer stmt.Close()

	status := 0
	ts := time.Now().Unix()

	res, err := stmt.Exec(0, schoolType, schoolName, country, province, city, latitude, longitude, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	schoolId, err := res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
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
		log.Error(err.Error())
		return
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

	log.Debugf("end AddSchool")
	return
}
