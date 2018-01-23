package account

import (
	"common/constant"
	"common/mymgo"
	"common/rest"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
	"svr/st"
)

type SearchSchoolsReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	SrcType int    `schema:"srcType"` // 0从ES 1从mgo

	SchoolName string `schema:"schoolName"`
	SchoolType int    `schema:"schoolType"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type SearchSchoolsRsp struct {
	Schools []*st.SchoolInfo2 `json:"schools"`
}

type EsSearchRspBody struct {
	Took uint32       `json:"took"`
	Hits EsSearchHits `json:"hits"`
}

type EsSearchHits struct {
	Total    uint32               `json:"total"`
	MaxScore float64              `json:"max_score"`
	Hits     []EsSearchHitElement `json:"hits"`
}

type EsSearchHitElement struct {
	Index  string         `json:"_index"`
	Type   string         `json:"_type"`
	Id     string         `json:"_id"`
	Score  float64        `json:"_score"`
	Source st.SchoolInfo2 `json:"_source"`
}

/*
type EsElementSource struct {
	SchoolId   uint32 `json:"schoolId"`
	SchoolType string `json:"schoolType"`
	SchoolName string `json:"schoolName"`
	Country    string `json:"country"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Status     uint32 `json:"status"`
	Ts         uint32 `json:"ts"`
}
*/

func doSearchSchools(req *SearchSchoolsReq, r *http.Request) (rsp *SearchSchoolsRsp, err error) {

	log.Debugf("uin %d, SearchSchoolsReq %+v", req.Uin, req)

	var schools []*st.SchoolInfo2

	if req.SrcType == 0 {

		schools, err = SearchSchoolsByNameFromEs(req.Uin, req.SchoolType, req.SchoolName, req.PageNum, req.PageSize)
		if err != nil {
			log.Errorf("uin %d, SearchSchoolsRsp error, %s", req.Uin, err.Error())
			return
		}
	} else {

		schools, err = SearchSchoolsByNameFromMgo(req.Uin, req.SchoolType, req.SchoolName, req.PageNum, req.PageSize)
		if err != nil {
			log.Errorf("uin %d, SearchSchoolsRsp error, %s", req.Uin, err.Error())
			return
		}
	}

	rsp = &SearchSchoolsRsp{schools}

	log.Debugf("uin %d, SearchSchoolsRsp succ, %+v", req.Uin, rsp)

	return
}

func SearchSchoolsByNameFromMgo(uin int64, schoolType int, name string, pageNum, pageSize int) (schools []*st.SchoolInfo2, err error) {

	schoolName := strings.TrimSpace(name)

	schools = make([]*st.SchoolInfo2, 0)

	if uin == 0 || len(schoolName) == 0 || schoolType > constant.ENUM_SCHOOL_TYPE_UNIVERSITY || schoolType < 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
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

	regName := fmt.Sprintf(".*%s.*", schoolName)

	if schoolType == 0 {
		err = c.Find(bson.M{"name": bson.RegEx{regName, ""}}).Limit(100).All(&mgoSchools)
	} else {
		err = c.Find(bson.M{"name": bson.RegEx{regName, ""}, "type": schoolType}).Limit(100).All(&mgoSchools)
	}

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

func SearchSchoolsByNameFromEs(uin int64, schoolType int, name string, pageNum, pageSize int) (schools []*st.SchoolInfo2, err error) {

	schoolName := strings.TrimSpace(name)

	schools = make([]*st.SchoolInfo2, 0)

	if uin == 0 || len(schoolName) == 0 || schoolType > constant.ENUM_SCHOOL_TYPE_UNIVERSITY || schoolType < 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	url := fmt.Sprintf(`http://122.152.206.97:9200/yplay/schools/_search`)

	s := fmt.Sprintf(`
		{
		    "query":{
		        "bool":{
		            "must": [
		                {"query_string": {"default_field":"school","query":"%s"}},
		                {"term":{"schoolType":"%d"}}		                		                
		            ]
		         }
		    },
		    "from":%d,
		    "size":%d,
		    "sort":[],
		    "aggs":{}
		}`, schoolName, schoolType, (pageNum-1)*pageSize, pageSize)

	if schoolType == 0 {
		s = fmt.Sprintf(`
		{
		    "query":{
		        "bool":{
		            "must": [
		                {"query_string": {"default_field":"school","query":"%s"}}		             		                		                
		            ]
		         }
		    },
		    "from":%d,
		    "size":%d,
		    "sort":[],
		    "aggs":{}
		}`, schoolName, (pageNum-1)*pageSize, pageSize)
	}

	rsp, err := http.Post(url, "application/json", strings.NewReader(s))
	if err != nil {
		return
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		log.Error("rsp StatusCode %d", rsp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Error("rsp body readall %s", err.Error())
		return
	}

	var ret EsSearchRspBody

	err = json.Unmarshal(body, &ret)
	if err != nil {
		log.Errorf("unmarshal err %s, body %s", err, string(body))
		return
	}

	for _, e := range ret.Hits.Hits {
		si := e.Source
		schools = append(schools, &si)
	}

	return
}
