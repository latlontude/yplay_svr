package experience

import (
	"api/elastSearch"
	"api/sns"
	"net/http"
	"regexp"
	"unicode"
)

//拉去最新经验贴 只展示名字

type SearchAllReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int    `schema:"boardId"`
	Content  string `schema:"content"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

type InterlocutionRsp struct {
	Interlocution []*elastSearch.Interlocution `json:"interlocutions"`
	TotalCnt      int                          `json:"totalCnt"`
}

type SearchAllRsp struct {
	Friends          []*sns.SearchFriendInfo `json:"friends"`
	GetLabelListRsp  GetLabelListRsp         `json:"getLabelListRsp"`
	InterlocutionRsp InterlocutionRsp        `json:"interlocutionRsp"`
}

func doSearchAll(req *SearchAllReq, r *http.Request) (rsp *SearchAllRsp, err error) {

	log.Debugf("uin %d, SearchAllReq succ, %+v", req.Uin, req)
	//TODO search  pupu用户       search label    search question and answer
	friends, getLabelListRsp, interlocutionRsp, err := SearchAll(req.Uin, req.BoardId, req.Content, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetLabelListReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SearchAllRsp{friends, getLabelListRsp, interlocutionRsp}

	log.Debugf("uin %d, SearchAllRsp succ, %+v", req.Uin, rsp)
	return
}

func SearchAll(uin int64, boardId int, content string, pageNum int, pageSize int) (friends []*sns.SearchFriendInfo, getLabelListRsp GetLabelListRsp, interlocutionRsp InterlocutionRsp, err error) {

	isCn := hasCn(content)
	if isCn == false {
		friends, err = sns.SearchFriends(uin, content)
		if err != nil {
			log.Debugf("get friends error ,friends:%v", friends)
		}
	}

	//查找属于该墙的label
	//labelList, labelListCnt, err := GetLabelInfoByBoardId(boardId, content, pageNum, pageSize)

	labelList, labelListCnt, err := elastSearch.SearchLabelFromEs(uin, content, boardId, pageNum, pageSize)
	if err != nil {
		log.Errorf("uin %d, GetLabelList error, %s", uin, err.Error())
		return
	}
	getLabelListRsp = GetLabelListRsp{labelList, labelListCnt}

	if labelListCnt >= pageSize {
		return
	} else {
		pageSize = pageSize - labelListCnt
	}

	interlocution, totalCnt, err := elastSearch.SearchInterlocutionFromEs(uin, content, boardId, pageNum, pageSize)
	if err != nil {
		log.Errorf("uin %d, GetLabelList error, %s", uin, err.Error())
	}
	interlocutionRsp = InterlocutionRsp{interlocution, totalCnt}

	return
}

//匹配中文字符
func hasCn(content string) (isCn bool) {
	isCn = false
	for _, r := range content {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u4e00-\u9fa5]").MatchString(string(r))) {
			isCn = true
			break
		}
	}
	return
}
