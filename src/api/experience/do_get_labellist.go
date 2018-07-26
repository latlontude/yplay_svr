package experience

import (
	"api/label"
	"net/http"
)

type GetLabelListReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	//AnswerId   int    `schema:"answerId"`
	LabelName string `schema:"labelName"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetLabelListRsp struct {
	LabelList []*label.LabelInfo `json:"labelList"`
	TotalCnt  int                `json:"totalCnt"`
}

func doGetLabelList(req *GetLabelListReq, r *http.Request) (rsp *GetLabelListRsp, err error) {

	log.Debugf("uin %d, GetLabelListReq %+v", req.Uin, req)

	labelList, totalCnt, err := label.GetLabelList(req.Uin, req.LabelName, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetLabelListReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetLabelListRsp{labelList, totalCnt}

	log.Debugf("uin %d, PostLikeRsp succ, %+v", req.Uin, rsp)

	return
}
