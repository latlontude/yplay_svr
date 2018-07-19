package elastSearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type ESAnswerRspBody struct {
	Took uint32       `json:"took"`
	Hits ESAnswerHits `json:"hits"`
}

type ESAnswerHits struct {
	Total    uint32               `json:"total"`
	MaxScore float64              `json:"max_score"`
	Hits     []ESAnswerHitElement `json:"hits"`
}

type ESAnswerHitElement struct {
	Index    string   `json:"_index"`
	Type     string   `json:"_type"`
	Id       string   `json:"_id"`
	Score    float64  `json:"_score"`
	EsAnswer EsAnswer `json:"_source"`
}

type EsAnswer struct {
	Qid           int    `json:"qid"`
	AnswerId      int    `json:"answerId"`
	AnswerContent string `json:"answerContent"`
}

func AddAnswerToEs(qid, answerId int, answerContent string) (err error) {

	var answer EsAnswer
	answer.Qid = qid
	answer.AnswerId = answerId
	answer.AnswerContent = answerContent

	d, err := json.Marshal(&answer)
	client := &http.Client{}
	url := fmt.Sprintf("http://122.152.206.97:9200/interlocution/answers/%d", answerId)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(d))
	if err != nil {
		log.Debugf(err.Error())
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
