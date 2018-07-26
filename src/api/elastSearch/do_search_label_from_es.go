package elastSearch

import (
	"api/label"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type SearchLabelFromEsReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	Content  string `schema:"content"`
	BoardId  int    `schema:"boardId"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

type SearchLabelRsp struct {
	LabelList []*label.LabelInfo `json:"labelList"`
	TotalCnt  int                `json:"totalCnt"`
}

func doSearchLabelFromEs(req *SearchLabelFromEsReq, r *http.Request) (rsp *SearchLabelRsp, err error) {

	log.Debugf("uin %d, SearchLabelFromEsReq %+v", req.Uin, req)

	LabelList, totalCnt, err := SearchLabelFromEs(req.Uin, req.Content, req.BoardId, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, SearchLabelRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SearchLabelRsp{LabelList, totalCnt}

	log.Debugf("uin %d, SearchInterlocutionFromEsRsp succ, %+v", req.Uin, rsp)

	return
}

func SearchLabelFromEs(uin int64, content string, boardId, pageNum, pageSize int) (labelList []*label.LabelInfo, totalCnt int, err error) {

	totalCnt = 0

	if uin == 0 {
		return
	}

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	hashLabel := make(map[int]label.LabelInfo, 0)

	url := fmt.Sprintf(`http://122.152.206.97:9200/interlocution/labels/_search`)
	s := fmt.Sprintf(`
	{
		"query":{"bool":{"must": [{"query_string":{"default_field":"labelName","query":"%s"}},{"term":{"boardId":%d}}]}},
		"from":%d,
		"size":%d,
		"sort":[],
		"aggs":{}
	}`, content, boardId, (pageNum-1)*pageSize, pageSize)

	log.Debugf("url :%s ,query string :%s", url, s)

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

	var esLabelRspBody ESLabelRspBody

	err = json.Unmarshal(body, &esLabelRspBody)
	if err != nil {
		log.Errorf("unmarshal err %s, body %s", err, string(body))
		return
	}

	for _, e := range esLabelRspBody.Hits.Hits {
		si := e.EsLabel
		var labelInfo label.LabelInfo
		labelInfo.LabelId = si.LabelId
		labelInfo.LabelName = si.LabelName

		//去重
		if _, ok := hashLabel[si.LabelId]; ok {
			continue
		} else {
			hashLabel[si.LabelId] = labelInfo
			labelList = append(labelList, &labelInfo)
		}
	}

	return
}
