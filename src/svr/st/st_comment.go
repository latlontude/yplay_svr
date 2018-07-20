package st

import (
	"fmt"
)

type ReplyInfo struct {
	ReplyId           int              `json:"replyId"`
	ReplyContent      string           `json:"replyContent"`
	ReplyFromUserInfo *UserProfileInfo `json:"fromUserInfo"`
	ReplyToUserInfo   *UserProfileInfo `json:"toUserInfo"`
	ReplyTs           int64            `json:"replyTs"`
	LikeCnt           int              `json:"likeCnt"`
	IsILike           bool             `json:"isILike"`
}

type CommentInfo struct {
	AnswerId       int              `json:"answerId"`
	CommentId      int              `json:"commentId"`
	CommentContent string           `json:"commentContent"`
	OwnerInfo      *UserProfileInfo `json:"ownerInfo"`
	CommentTs      int              `json:"commentTs"`
	Replys         []ReplyInfo      `json:"replys"`
	LikeCnt        int              `json:"likeCnt"`
	IsILike        bool             `json:"isILike"`
}

func (this *ReplyInfo) String() string {

	return fmt.Sprintf(`ReplyInfo{ReplyId:%d ReplyContent：%s ReplyFromUserInfo:%+v ReplyToUserInfo:%+v ReplyTs:%d LikeCnt:%d IsILike:%t}`,
		this.ReplyId, this.ReplyContent, this.ReplyFromUserInfo, this.ReplyToUserInfo, this.ReplyTs, this.LikeCnt, this.IsILike)
}

func (this *CommentInfo) String() string {

	return fmt.Sprintf(`CommentInfo{AnswerId:%d CommentId:%d CommentContent:%s OwnerInfo:%+v CommentTs:%d Replys:%+v  LikeCnt:%d IsILike:%t}`,
		this.AnswerId, this.CommentId, this.CommentContent, this.OwnerInfo, this.CommentTs, this.Replys, this.LikeCnt, this.IsILike)
}
