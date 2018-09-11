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
	Question          st.V2QuestionInfo `json:"question"`          //问题
	Answers           st.AnswersInfo    `json:"answer"`            //若干回答
	QuestionHighlight string            `json:"questionHighlight"` //高亮
	AnswerHighlight   string            `json:"answerHighlight"`   //高亮
	Type              int               `json:"type"`              //类型    3:提问和回答都被匹配  2:回答被匹配  1:提问被匹配
}

const (
	PRE_TAG    = "<K9j4A1cw0sV>"
	END_TAG    = "</K9j4A1cw0sV>"
	FONT_START = "<font color='#0092E9'>"
	FONT_END   = "</font>"

	QUESTION_HIGHLIGHT_LENGTH = 22
	ANSWER_HIGHLIGHT_LENGTH   = 63

	TYPE_QUESTION_HIGHLIGHT = 0
	TYPE_ANSWER_HIGHLIGHT   = 1
)

type SearchInterlocutionFromEsReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	Content  string `schema:"content"`
	BoardId  int    `schema:"boardId"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
	Version  int    `schema:"version"`
}

type SearchInterlocutionFromEsRsp struct {
	Interlocution []*Interlocution `json:"interlocutions"`
	TotalCnt      int              `json:"totalCnt"`
}

func doSearchInterlocutionFromEs(req *SearchInterlocutionFromEsReq, r *http.Request) (rsp *SearchInterlocutionFromEsRsp, err error) {

	log.Debugf("uin %d, SearchInterlocutionFromEsReq %+v", req.Uin, req)

	interlocution, totalCnt, err := SearchInterlocutionFromEs(req.Uin, req.Content, req.BoardId, req.PageNum, req.PageSize, req.Version)

	if err != nil {
		log.Errorf("uin %d, SearchInterlocutionFromEsReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SearchInterlocutionFromEsRsp{interlocution, totalCnt}

	log.Debugf("uin %d, SearchInterlocutionFromEsRsp succ, %+v", req.Uin, rsp)

	return
}

func SearchInterlocutionFromEs(uin int64, content string, boardId, pageNum, pageSize int, version int) (interlocution []*Interlocution, totalCnt int, err error) {

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
		"aggs":{},
		"highlight": {
            "pre_tags": ["%s"],
            "post_tags": ["%s"],
        	"fields" : {
            	"answerContent" : {"number_of_fragments":0}
        	}
    	}
	}`, content, boardId, (pageNum-1)*pageSize, pageSize, PRE_TAG, END_TAG)

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
		highlight := e.Highlight
		log.Debugf("boardId:%d answerId :%d qid :%d , answerContent : %s,high:%+v", answerBoardId, answerId, qid, answerContent, highlight)

		answer, err2 := common.GetV2Answer(answerId)
		if err2 != nil {
			log.Errorf("get answer error")
			continue
		}
		hashQidAnswerId[answerId] = qid

		var interAnswer Interlocution
		interAnswer.Type = 2
		interAnswer.Answers = answer
		//if uin == 103004 || uin == 103096 {
		//只取第一个
		content := highlight.HighlightContent[0]
		interAnswer.AnswerHighlight = GetHighlightString(content, TYPE_QUESTION_HIGHLIGHT)
		//}
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
		"aggs":{},
		"highlight": {
            "pre_tags": ["%s"],
            "post_tags": ["%s"],
        	"fields" : {
            	"qContent" : {"number_of_fragments":0}
        	}
    	}
	}`, content, boardId, (pageNum-1)*pageSize, pageSize, PRE_TAG, END_TAG)

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
		highlight := e.Highlight
		log.Debugf("boardId:%d,qid :%d , qContent : %s,high:%+v", qstBoardId, qid, qContent, highlight)

		question, err1 := common.GetV2Question(qid, version)
		if err1 != nil {
			log.Errorf("get question error")
			continue
		}

		for k, v := range hashQidAnswerId {
			if v == qid {
				qstAnswerIs = k
			}
		}

		var interQts Interlocution
		//说明之前匹配到回答
		if qstAnswerIs != 0 {
			interQts = hashmap[qstAnswerIs]
			interQts.Question = question
			interQts.Type = 3
			delete(hashmap, qstAnswerIs)
		} else {
			interQts.Question = question
			interQts.Type = 1
		}

		//if uin == 103004 || uin == 103096 {
		//只取第一个
		content := highlight.HighlightContent[0]
		interQts.QuestionHighlight = GetHighlightString(content, TYPE_ANSWER_HIGHLIGHT)
		//}
		interlocution = append(interlocution, &interQts)
		totalCnt++
	}

	//hash 剩余的(只有回答没有匹配到提问的)
	for k, v := range hashmap {
		qid := hashQidAnswerId[k]
		question, err2 := common.GetV2Question(int(qid), version)
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

func GetHighlightString(content string, highlightType int) (highlight string) {

	var runeString string

	emStart := PRE_TAG
	emEnd := END_TAG
	start := strings.Index(content, emStart)
	end := strings.Index(content, emEnd) + len(emEnd)

	//tag 字符总数
	labelLength := len([]rune(FONT_START)) + len([]rune(FONT_END))

	//字符总个数
	totalLength := len([]rune(content)) - labelLength

	log.Debugf("content:%s,totalLength:%d", content, totalLength)

	showLength := 0
	if highlightType == 0 {
		showLength = QUESTION_HIGHLIGHT_LENGTH
	} else {
		showLength = ANSWER_HIGHLIGHT_LENGTH
	}

	if totalLength > showLength {
		//匹配到的字符个数
		highLength := len([]rune(content[start:end]))

		//剩余字符数
		restLength := showLength + labelLength - highLength - 6

		//剩余字符一半
		half := restLength / 2

		//左半部字符串
		beforeString := content[:start]
		//右半部字符串
		endString := content[end:]

		beforeRune := []rune(beforeString)
		beforeRuneLength := len(beforeRune)

		endRune := []rune(endString)
		endRuneLength := len(endRune)

		var left, right []rune
		if beforeRuneLength > half {
			if endRuneLength > half {
				left = beforeRune[beforeRuneLength-half:]
				right = endRune[:half]
			} else {
				right = endRune
				//剩余全是左边的
				leftTotal := restLength - endRuneLength

				log.Debugf("restlen:%d,leftTotal:%d,beforeLen:%d,endLen:%d,content:%s",
					restLength, leftTotal, beforeRuneLength, endRuneLength, content)
				if leftTotal > beforeRuneLength {
					left = beforeRune
				} else {
					//取left  + harf - right
					left = beforeRune[beforeRuneLength-leftTotal:]
				}
				log.Debugf("left:%s right:%s", string(left), string(right))
			}
		} else {
			if beforeRuneLength > half {
				left = beforeRune[beforeRuneLength-half:]
				right = endRune[:half]
				log.Debugf("left:%s right:%s", string(left), string(right))
			} else {
				left = beforeRune
				//剩余全是右边的
				rightTotal := restLength - beforeRuneLength
				if rightTotal > endRuneLength {
					right = endRune
				} else {
					//取left  + harf - right
					right = endRune[:rightTotal]
				}
				log.Debugf("left:%s right:%s", string(left), string(right))
			}
		}
		log.Debugf("left:%s right:%s , content:%s , part:%s", string(left), string(right), content, content[start:end])

		runeString = string("...") + string(left) + content[start:end] + string(right) + string("...")

	} else {
		runeString = content
	}

	runeString = strings.Replace(runeString, PRE_TAG, FONT_START, -1)
	runeString = strings.Replace(runeString, END_TAG, FONT_END, -1)

	highlight = runeString

	log.Debugf("runeString:%s", runeString)
	return
}
