package experience

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"svr/st"
)

//type ExpLabel struct {
//	LabelId     int                 `json:"labelId"`
//	LabelName   string              `json:"labelName"`
//}

////经验谈 标签的信息
type ExperienceInfo struct {
	Question   st.V2QuestionInfo `json:"question"`
	AnswerInfo st.AnswersInfo    `json:"answer"`
	Ts         int64             `json:"ts"`
}

var (
	APIMap = httputil.APIMap{
		"/getLabelList":       auth.Apify2(doGetLabelList), //获取标签列表
		"/getExpDetail":       auth.Apify2(doGetExpDetail), //经验贴详情 列出带该标签的所有问题
		"/addAnswerIdInExp":   auth.Apify2(doAddAnswerIdInExp),
		"/delAnswerIdFromExp": auth.Apify2(doDelAnswerIdFromExp),
		"/getExpHome":         auth.Apify2(doGetExpHome),
		"/searchAll":          auth.Apify2(doSearchAll),
	}

	log = env.NewLogger("experience")
)
