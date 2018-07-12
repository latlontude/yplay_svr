package experience

import (
	"api/question"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"sort"
	"svr/st"
)

// 自定义排序
type quest []*st.V2QuestionInfo

func (I quest) Len() int {
	return len(I)
}

func (I quest) Less(i, j int) bool {
	return I[i].CreateTs > I[j].CreateTs
}

func (I quest) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

//拉去最新经验贴 只展示名字

type GetExperienceDetailReq struct {
	Uin         int64  `schema:"uin"`
	Token       string `schema:"token"`
	Ver         int    `schema:"ver"`

	BoardId     int    `schema:"boardId"`
	LabelId     int    `schema:"labelId"`
	PageNum     int    `schema:"pageNum"`
	PageSize    int    `schema:"pageSize"`

}

type GetExperienceDetailRsp struct {
	Questions  []*st.V2QuestionInfo          `json:"questions"`
	TotalCnt  int                            `json:"totalCnt"`
}

func doGetExperienceDetail(req *GetExperienceDetailReq, r *http.Request) (rsp *GetExperienceDetailRsp, err error) {

	log.Debugf("uin %d, GetExperienceDetailReq succ, %+v", req.Uin, rsp)

	questions,totalCnt,err := GetExperienceDetail(req.Uin,req.BoardId,req.LabelId,req.PageNum,req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetExperienceDetailReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetExperienceDetailRsp{questions,totalCnt}

	log.Debugf("uin %d, GetExperienceDetailRsp succ, %+v", req.Uin, rsp)

	return
}


/**
		labelId labelName count updateTs qid question
 */

func GetExperienceDetail(uin int64, boardId , labelId , pageNum , pageSize int) (questions []*st.V2QuestionInfo, totalCnt int, err error){


	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	start   := (pageNum -1) * pageSize
	end     := start + pageSize

	totalCnt = 0
	questions = make([]*st.V2QuestionInfo, 0)

	sql := fmt.Sprintf(`select experience_share.qid,experience_share.ts ,
			v2questions.ownerUid, v2questions.qTitle, v2questions.qContent, v2questions.qImgUrls, 
			v2questions.isAnonymous, v2questions.createTs, v2questions.modTs 
			from v2questions,experience_share 
			where experience_share.boardId = %d and experience_share.labelId = %d 
			and (v2questions.qContent != "" or v2questions.qImgUrls != "")
			and experience_share.qid = v2questions.qid limit %d,%d`,
				boardId,labelId,start,end)

	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var info st.V2QuestionInfo
		var uid int64
		var createTs int64

		rows.Scan(
			&info.Qid,
			&info.CreateTs,
			&uid,
			&info.QTitle,
			&info.QContent,
			&info.QImgUrls,
			&info.IsAnonymous,
			&createTs,
			&info.ModTs)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}

			info.OwnerInfo = ui
		}

		answerCnt, err := question.GetAnswerCnt(info.Qid)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		info.AnswerCnt = answerCnt

		bestAnswer, _ := question.GetBestAnswer(uin, info.Qid)
		info.BestAnswer = bestAnswer

		responders, _ := question.GetQidNewResponders(info.Qid)
		info.NewResponders = responders

		questions = append(questions, &info)

		totalCnt++
	}

	//排序
	sort.Sort(quest(questions))

	return

}