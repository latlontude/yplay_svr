package st

import (
	"fmt"
)

type AnswersInfo struct {
	Qid           int              `json:"qid"`
	AnswerId      int              `json:"answerId"`
	AnswerContent string           `json:"answerContent"`
	AnswerImgUrls string           `json:"answerImgUrls"`
	AnswerTs      int              `json:"answerTs"`
	CommentCnt    int              `json:"commentCnt"`
	LikeCnt       int              `json:"likeCnt"`
	IsILike       bool             `json:"isILike"`
	OwnerInfo     *UserProfileInfo `json:"ownerInfo"`
}

func (this *AnswersInfo) String() string {

	return fmt.Sprintf(`AnswersInfo{Qid:%d AnswerId:%d AnswerContent:%s AnswerImgUrls:%s AnswerTs:%d CommentCnt:%d LikeCnt:%d IsILike:%t OwnerInfo:%+v}`,
		this.Qid, this.AnswerId, this.AnswerContent, this.AnswerImgUrls, this.AnswerTs, this.CommentCnt, this.LikeCnt, this.IsILike, this.OwnerInfo)
}
