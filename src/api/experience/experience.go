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
		"/getBoardLabel":      auth.Apify2(doGetExpHome), //主页 运营可配置
		"/searchAll":          auth.Apify2(doSearchAll),  //搜索  pupu用户 经验弹 问题和回答
		"/addAdmin":           auth.Apify2(doAddAdmin),
		"/getAdminList":       auth.Apify2(doGetAdminList),
		"/addLabel":           auth.Apify2(doAddLabel),
	}

	log = env.NewLogger("experience")
)

//curl http://122.152.206.97:9200/interlocution/questions/_search -d '
//	{
//		"query":{"bool":{"must": [{"query_string":{"default_field":"qContent","query":"哈哈"}},{"term":{"boardId":4}}]}},
//		"from":0,
//		"size":20,
//		"sort":[],
//		"aggs":{}
//	}
//	'
