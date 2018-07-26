package elastSearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type ESLabelRspBody struct {
	Took uint32      `json:"took"`
	Hits ESLabelHits `json:"hits"`
}

type ESLabelHits struct {
	Total    uint32              `json:"total"`
	MaxScore float64             `json:"max_score"`
	Hits     []ESLabelHitElement `json:"hits"`
}

type ESLabelHitElement struct {
	Index   string  `json:"_index"`
	Type    string  `json:"_type"`
	Id      string  `json:"_id"`
	Score   float64 `json:"_score"`
	EsLabel EsLabel `json:"_source"`
}

type EsLabel struct {
	BoardId   int    `json:"boardId"`
	LabelId   int    `json:"labelId"`
	LabelName string `json:"labelName"`
}

func AddLabelToEs(boardId, answerId, labelId int, labelName string) (err error) {

	var label EsLabel
	label.BoardId = boardId
	label.LabelId = labelId
	label.LabelName = labelName

	d, err := json.Marshal(&label)
	client := &http.Client{}
	url := fmt.Sprintf("http://122.152.206.97:9200/interlocution/labels/%d", answerId)
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

	log.Debugf("dump:%s , req:%v", string(dump), req)
	return
}

func DelLabelToEs(answerId int) (err error) {

	client := &http.Client{}
	url := fmt.Sprintf("http://122.152.206.97:9200/interlocution/labels/%d", answerId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	client.Do(req)
	req.Header.Add("Content-Type", "application/json")
	dump, err := httputil.DumpRequest(req, true)

	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Debugf("dump:%s , req:%v", string(dump), req)
	return
}
