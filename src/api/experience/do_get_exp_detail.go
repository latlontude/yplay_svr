package experience

import (
	"api/common"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"sort"
	"svr/st"
)

// 自定义排序
type Interface []*ExperienceInfo

func (I Interface) Len() int {
	return len(I)
}

func (I Interface) Less(i, j int) bool {
	return I[i].Ts > I[j].Ts
}

func (I Interface) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

//拉去最新经验贴 只展示名字

type GetExperienceDetailReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	BoardId  int `schema:"boardId"`
	LabelId  int `schema:"labelId"`
	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type GetExperienceDetailRsp struct {
	ExpInfo    []*ExperienceInfo     `json:"expInfo"`
	TotalCnt   int                   `json:"totalCnt"`
	Operators  []*st.UserProfileInfo `json:"operators"`
	UpdateTime int64                 `json:"updateTime"`
}

func doGetExpDetail(req *GetExperienceDetailReq, r *http.Request) (rsp *GetExperienceDetailRsp, err error) {

	log.Debugf("uin %d, GetExperienceDetailReq succ, %+v", req.Uin, req)

	expInfo, totalCnt, operators, updateTime, err := getExpDetail(req.Uin, req.BoardId, req.LabelId, req.PageNum, req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetExperienceDetailReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetExperienceDetailRsp{expInfo, totalCnt, operators, updateTime}

	log.Debugf("uin %d, GetExperienceDetailRsp succ, %+v", req.Uin, rsp)

	return
}

/**
labelId labelName count updateTs qid question
*/

func getExpDetail(uin int64, boardId, labelId, pageNum, pageSize int) (ExpInfo []*ExperienceInfo, totalCnt int, operators []*st.UserProfileInfo, updateTime int64, err error) {

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

	start := (pageNum - 1) * pageSize
	end := start + pageSize

	totalCnt = 0
	ExpInfo = make([]*ExperienceInfo, 0)

	sql := fmt.Sprintf(`select  qid,answerId,ts from experience_share where boardId = %d  and labelId = %d limit %d,%d`, boardId, labelId, start, end)

	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var Exp ExperienceInfo
		var qid, answerId int
		var ts int64

		rows.Scan(&qid, &answerId, &ts)

		Exp.Question, _ = common.GetV2Question(qid)
		Exp.AnswerInfo, _ = common.GetV2Answer(answerId)

		//点赞数

		likeCnt, err1 := common.GetAnswerLikeCnt(answerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		Exp.AnswerInfo.LikeCnt = likeCnt
		isILike, err1 := common.CheckIsILikeAnswer(uin, answerId)
		if err1 != nil {
			log.Error(err1.Error())
			continue
		}

		Exp.AnswerInfo.IsILike = isILike

		Exp.Ts = ts
		ExpInfo = append(ExpInfo, &Exp)
		totalCnt++
	}

	//排序
	sort.Sort(Interface(ExpInfo))

	//updateTime = ExpInfo[0].Ts

	//整理过经验弹的人  找到最新时间
	sql = fmt.Sprintf(`select operator,max(ts) from experience_share where boardId = %d and labelId = %d  group by operator order by ts`, boardId, labelId)

	rows, err = inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	for rows.Next() {
		var uid int64

		rows.Scan(&uid, &updateTime)

		if uid > 0 {
			ui, err1 := st.GetUserProfileInfo(uid)
			if err1 != nil {
				log.Error(err1.Error())
				continue
			}
			operators = append(operators, ui)
		}
	}

	return

}
