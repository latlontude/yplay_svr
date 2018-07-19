package elastSearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type ESQuestionRspBody struct {
	Took uint32         `json:"took"`
	Hits ESQuestionHits `json:"hits"`
}

type ESQuestionHits struct {
	Total    uint32                 `json:"total"`
	MaxScore float64                `json:"max_score"`
	Hits     []ESQuestionHitElement `json:"hits"`
}

type ESQuestionHitElement struct {
	Index      string     `json:"_index"`
	Type       string     `json:"_type"`
	Id         string     `json:"_id"`
	Score      float64    `json:"_score"`
	EsQuestion EsQuestion `json:"_source"`
}

type EsQuestion struct {
	Qid      int    `json:"qid"`
	QContent string `json:"qContent"`
}

func AddQstToEs(qid int, qContent string) (err error) {

	var question EsQuestion
	question.Qid = qid
	question.QContent = qContent

	d, err := json.Marshal(&question)
	client := &http.Client{}
	url := fmt.Sprintf("http://122.152.206.97:9200/interlocution/questions/%d", qid)
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
