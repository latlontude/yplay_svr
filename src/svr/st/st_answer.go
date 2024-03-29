package st

import (
	"fmt"
)

type ExpLabel struct {
	LabelId   int    `json:"labelId"`
	LabelName string `json:"labelName"`
}

type AnswersInfo struct {
	Qid           int              `json:"qid"`
	AnswerId      int              `json:"answerId"`
	AnswerContent string           `json:"answerContent"`
	AnswerImgUrls string           `json:"answerImgUrls"`
	AnswerTs      int              `json:"answerTs"`
	Ext           string           `json:"ext"`
	CommentCnt    int              `json:"commentCnt"`
	LikeCnt       int              `json:"likeCnt"`
	IsILike       bool             `json:"isILike"`
	OwnerInfo     *UserProfileInfo `json:"ownerInfo"`
	ExpLabel      []*ExpLabel      `json:"expLabel"`
	LatestComment []*CommentInfo   `json:"latestComment"`
	IsAnonymous   bool             `json:"isAnonymous"`
	Reply         []*CommentInfo   `json:"reply"`
}

func (this *AnswersInfo) String() string {

	return fmt.Sprintf(`AnswersInfo{Qid:%d AnswerId:%d AnswerContent:%s AnswerImgUrls:%s AnswerTs:%d CommentCnt:%d LikeCnt:%d IsILike:%t OwnerInfo:%+v}`,
		this.Qid, this.AnswerId, this.AnswerContent, this.AnswerImgUrls, this.AnswerTs, this.CommentCnt, this.LikeCnt, this.IsILike, this.OwnerInfo)
}
