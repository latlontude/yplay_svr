package question

import (
	"api/common"
	"common/constant"
	"common/mymgo"
	"common/rest"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"svr/st"
)

type GeoJson struct {
	Type        string    `json:"string"`
	Coordinates []float64 `json:"coordinates"`
}

type MgoQuestionInfo struct {
	Qid      int     `bson:"_id,omitempty" json:"qid"`
	PoiTag   string  `bson:"poiTag" json:"poiTag"`
	Loc      GeoJson `bson:"loc"  json:"loc"`
	CreateTs int64   `bson:"createTs" json:"createTs"`
}

func GetNearbyQuestions(uin int64, boardId int, latitude, longitude float64, version int) (questions []*st.V2QuestionInfo, err error) {

	questions = make([]*st.V2QuestionInfo, 0)

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

	c := db.C("questions")
	if c == nil {
		err = rest.NewAPIError(constant.E_MGO_COLLECTION_NIL, "mgo collection nil")
		log.Error(err.Error())
		return
	}

	questionMgoMap := make([]MgoQuestionInfo, 0)

	err = c.Find(bson.M{
		"loc": bson.M{
			"$near": bson.M{
				"type":        "Point",
				"coordinates": []float64{longitude, latitude}}},
	}).Limit(1000).All(&questionMgoMap)

	if err != nil {
		err = rest.NewAPIError(constant.E_MGO_QUERY, err.Error())
		log.Error(err.Error())
		return
	}

	qidList := make([]int, 0)

	//TODO 获取未读的qid
	for _, mgoQst := range questionMgoMap {
		qidList = append(qidList, mgoQst.Qid)
	}

	return
}

func GetPoiTagQuestions(uin int64, boardId int, latitude, longitude float64, poiTag string, version int) (questions []*st.V2QuestionInfo, err error) {
	questions = make([]*st.V2QuestionInfo, 0)

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

	c := db.C("questions")
	if c == nil {
		err = rest.NewAPIError(constant.E_MGO_COLLECTION_NIL, "mgo collection nil")
		log.Error(err.Error())
		return
	}

	questionMgoMap := make([]MgoQuestionInfo, 0)

	err = c.Find(bson.M{
		"loc": bson.M{
			"$near": bson.M{
				"type":        "Point",
				"coordinates": []float64{longitude, latitude}}},
		"poiTag": poiTag}).Limit(1000).All(&questionMgoMap)

	if err != nil {
		err = rest.NewAPIError(constant.E_MGO_QUERY, err.Error())
		log.Error(err.Error())
		return
	}

	qidList := make([]int, 0)

	qidStr := ""
	for _, mgoQst := range questionMgoMap {
		qidList = append(qidList, mgoQst.Qid)
		qidStr += fmt.Sprintf("%d,", mgoQst.Qid)
	}

	questions, err3 := common.GetQuestionsByQidStr(qidStr, version)
	if err3 != nil {
		return
	}

	return
}

func AddQuestionToMgo(uin int64, qid int, latitude, longitude float64, poiTag string, createTs int64) (err error) {

	//将学校信息存储到mgo中
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

	c := db.C("questions")
	if c == nil {
		err = rest.NewAPIError(constant.E_MGO_COLLECTION_NIL, "mgo collection nil")
		log.Error(err.Error())
		return
	}

	questionMgoMap := make(map[string]interface{})
	mgoLocalINfoMap := make(map[string]interface{})

	mgoLocalINfoMap["type"] = "Point"
	mgoLocalINfoMap["coordinates"] = []float64{longitude, latitude}

	questionMgoMap["_id"] = qid
	questionMgoMap["poiTag"] = poiTag
	questionMgoMap["loc"] = mgoLocalINfoMap
	questionMgoMap["createTs"] = createTs

	log.Debugf("questionMgoMap %+v", questionMgoMap)

	err = c.Insert(&questionMgoMap)
	if err != nil {
		log.Error(err.Error())
		return
	}
	return
}
