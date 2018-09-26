package question

import (
	"net/http"
	"svr/st"
)

type GetNearByQuestionsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId   int     `schema:"boardId"`
	Longitude float64 `schema:"longitude"`
	Latitude  float64 `schema:"latitude"`
	Radius    int     `schema:"radius"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
	Version  int `schema:"version"`
}

type GetNearByQuestionsRsp struct {
	V2Questions []*st.V2QuestionInfo `json:"questions"`
	TotalCnt    int                  `json:"totalCnt"`
}

func doGetNearbyQuestions(req *GetGeoQuestionsReq, r *http.Request) (rsp *GetGeoQuestionsRsp, err error) {

	log.Debugf("uin %d, GetQuestionsReq %+v", req.Uin, req)

	questions, totalCnt, err := GetGeoQuestions(req.Uin, req.Qid, req.BoardId, req.Longitude, req.Latitude, req.Radius, req.PoiTag, req.PageNum, req.PageSize, req.Version)

	if err != nil {
		log.Errorf("uin %d, GetQuestions error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetGeoQuestionsRsp{questions, totalCnt}

	log.Debugf("uin %d, GetQuestionsRsp succ  , rsp:%v", req.Uin, rsp)

	return
}
