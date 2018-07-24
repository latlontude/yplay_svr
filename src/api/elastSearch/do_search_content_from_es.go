package elastSearch

import (
	"api/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"svr/st"
)

// 自定义排序
type inter []*Interlocution

func (I inter) Len() int {
	return len(I)
}

func (I inter) Less(i, j int) bool {
	return I[i].Type > I[j].Type
}

func (I inter) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

type Interlocution struct {
	Question st.V2QuestionInfo `json:"question"` //问题
	Answers  st.AnswersInfo    `json:"answer"`   //若干回答
	Type     int               `json:"type"`     //类型    3:提问和回答都被匹配  2:回答被匹配  1:提问被匹配
}

type SearchInterlocutionFromEsReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	Content  string `schema:"content"`
	BoardId  int    `schema:"boardId"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

type SearchInterlocutionFromEsRsp struct {
	Interlocution []*Interlocution `json:"interlocutions"`
	TotalCnt      int              `json:"totalCnt"`
}

func doSearchInterlocutionFromEs(req *SearchInterlocutionFromEsReq, r *http.Request) (rsp *SearchInterlocutionFromEsRsp, err error) {

	log.Debugf("uin %d, SearchInterlocutionFromEsReq %+v", req.Uin, req)

	interlocution, totalCnt, err := SearchInterlocutionFromEs(req.Uin, req.Content, req.BoardId, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, SearchInterlocutionFromEsReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SearchInterlocutionFromEsRsp{interlocution, totalCnt}

	log.Debugf("uin %d, SearchInterlocutionFromEsRsp succ, %+v", req.Uin, rsp)

	return
}

func SearchInterlocutionFromEs(uin int64, content string, boardId, pageNum, pageSize int) (interlocution []*Interlocution, totalCnt int, err error) {

	totalCnt = 0

	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	hashmap := make(map[int]Interlocution, 0)

	//先搜回答
	url := fmt.Sprintf(`http://122.152.206.97:9200/interlocution/answers/_search`)
	s := fmt.Sprintf(`
	{
		"query":{"bool":{"must": [{"query_string":{"default_field":"answerContent","query":"%s"}},{"term":{"boardId":%d}}]}},
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

	var esAnswerRspBody ESAnswerRspBody

	err = json.Unmarshal(body, &esAnswerRspBody)
	if err != nil {
		log.Errorf("unmarshal err %s, body %s", err, string(body))
		return
	}

	hashQidAnswerId := make(map[int]int)
	for _, e := range esAnswerRspBody.Hits.Hits {
		si := e.EsAnswer
		answerId := si.AnswerId
		qid := si.Qid
		answerContent := si.AnswerContent
		answerBoardId := si.BoardId

		log.Debugf("boardId:%d answerId :%d qid :%d , answerContent : %s", answerBoardId, answerId, qid, answerContent)

		answer, err2 := common.GetV2Answer(answerId)
		if err2 != nil {
			log.Errorf("get answer error")
			continue
		}
		hashQidAnswerId[answerId] = qid

		var interAnswer Interlocution
		interAnswer.Type = 2
		interAnswer.Answers = answer
		hashmap[answerId] = interAnswer
	}

	//TODO question

	url = fmt.Sprintf(`http://122.152.206.97:9200/interlocution/questions/_search`)

	s = fmt.Sprintf(`
	{
		"query":{"bool":{"must": [{"query_string":{"default_field":"qContent","query":"%s"}},{"term":{"boardId":%d}}]}},
		"from":%d,
		"size":%d,
		"sort":[],
		"aggs":{}
	}`, content, boardId, (pageNum-1)*pageSize, pageSize)

	log.Debugf("url:%s,query string :%s", url, s)

	rsp, err = http.Post(url, "application/json", strings.NewReader(s))

	if err != nil {
		return
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		log.Error("rsp StatusCode %d", rsp.StatusCode)
		return
	}

	body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Error("rsp body readall %s", err.Error())
		return
	}

	var esQuestionRspBody ESQuestionRspBody

	err = json.Unmarshal(body, &esQuestionRspBody)
	if err != nil {
		log.Errorf("unmarshal err %s, body %s", err, string(body))
		return
	}

	for _, e := range esQuestionRspBody.Hits.Hits {
		qstAnswerIs := 0
		si := e.EsQuestion
		qid := si.Qid
		qContent := si.QContent
		qstBoardId := si.BoardId
		log.Debugf("boardId:%d,qid :%d , qContent : %s", qstBoardId, qid, qContent)

		question, err1 := common.GetV2Question(qid)
		if err1 != nil {
			log.Errorf("get question error")
			continue
		}

		for k, v := range hashQidAnswerId {
			if v == qid {
				qstAnswerIs = k
			}
		}

		//说明之前匹配到回答
		if qstAnswerIs != 0 {
			interQts := hashmap[qstAnswerIs]
			interQts.Question = question
			interQts.Type = 3
			delete(hashmap, qstAnswerIs)
			interlocution = append(interlocution, &interQts)
		} else {
			var interQts Interlocution
			interQts.Question = question
			interQts.Type = 1
			interlocution = append(interlocution, &interQts)
		}
		totalCnt++
	}

	//hash 剩余的(只有回答没有匹配到提问的)
	for k, v := range hashmap {
		qid := hashQidAnswerId[k]
		question, err2 := common.GetV2Question(int(qid))
		interQts := hashmap[k]
		interQts.Question = question
		if err2 != nil {
			continue
		}
		log.Debugf("k:%d qid:%d question:%v", k, qid, question)
		v.Question = question
		interlocution = append(interlocution, &interQts)
		totalCnt++
	}

	//排序
	sort.Sort(inter(interlocution))

	return
}
