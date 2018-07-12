package experience

import (
	"api/question"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)


//拉去最新经验贴 只展示名字

type GetExperienceLabelReq struct {
	Uin         int64  `schema:"uin"`
	Token       string `schema:"token"`
	Ver         int    `schema:"ver"`

	BoardId     int    `schema:"boardId"`
	PageNum     int    `schema:"pageNum"`
	PageSize    int    `schema:"pageSize"`

}

type GetExperienceLabelRsp struct {
	ExperienceInfo  []*ExperienceInfo   `json:"experienceInfo"`
	TotalCnt  int                       `json:"totalCnt"`
}

func doGetExperienceLabel(req *GetExperienceLabelReq, r *http.Request) (rsp *GetExperienceLabelRsp, err error) {

	log.Debugf("uin %d, GetExperienceLabelReq succ, %+v", req.Uin, rsp)

	experienceInfo,totalCnt,err := GetExperienceLabel(req.BoardId,req.PageNum,req.PageSize)

	if err != nil {
		log.Errorf("uin %d, GetExperienceLabelReq error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetExperienceLabelRsp{experienceInfo,totalCnt}

	log.Debugf("uin %d, GetExperienceLabelRsp succ, %+v", req.Uin, rsp)

	return
}


/**
		labelId labelName count updateTs qid question
 */

func GetExperienceLabel(boardId , pageNum ,pageSize int) (experienceInfo []*ExperienceInfo, totalCnt int, err error){


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

	sql := fmt.Sprintf(`select  experience_share.labelId , experience_label.labelName,  count(experience_share.labelId)as count ,
			experience_share.qid,experience_share.ts 
			from  experience_label ,experience_share 
			where experience_label.labelId=experience_share.labelId 
			and experience_share.boardId = %d
			group by experience_share.labelId 
			order by ts desc limit %d,%d`,boardId,start,end)


	log.Debugf("sql : %s",sql)

	rows, err := inst.Query(sql)
	defer rows.Close()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}

	totalCnt = 0
	for rows.Next() {
		var experienceTmp ExperienceInfo
		var qid int
		rows.Scan(&experienceTmp.LabelId,&experienceTmp.LabelName,&experienceTmp.QuestionCnt,&qid,&experienceTmp.UpdateTs)

		question,err1  := question.GetV2Question(qid)
		if err1 != nil {
			return
		}

		experienceTmp.Question = question
		experienceInfo = append(experienceInfo,&experienceTmp)

		totalCnt++
	}

	return

}