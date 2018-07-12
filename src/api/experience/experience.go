package experience

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"svr/st"
)

////经验谈 标签的信息
type ExperienceInfo struct{
	LabelId     int                 `json:"labelId"`
	LabelName   string              `json:"labelName"`
	Uin         int64               `json:"ownerUid"`
	CreateTs    int64               `json:"createTs"`
	UpdateTs    int64               `json:"updateTs"`
	QuestionCnt int                 `json:"questionCnt"`
	Question    st.V2QuestionInfo   `json:"question"`
}



var (
	APIMap = httputil.APIMap{
		"/getlabellist"         : auth.Apify2(doGetLabelList),              //获取标签列表
		"/getexperience"        : auth.Apify2(doGetExperienceLabel),        //根据boardId获取经验帖 显示更新时间和一条问题
		"/getexperiencedetail"  : auth.Apify2(doGetExperienceDetail),       //经验贴详情 列出带该标签的所有问题
		"/addqidinexp"          : auth.Apify2(doAddQidInExperience),
		"/delqidinexp"          : auth.Apify2(doDelQidInExperience),
	}

	log = env.NewLogger("experience")
)