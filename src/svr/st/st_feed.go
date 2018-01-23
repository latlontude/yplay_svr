package st

import (
	"fmt"
)

type FeedInfo struct {
	VoteRecordId int64 `json:"voteRecordId"`

	FriendUin        int64  `json:"friendUin"`
	FriendNickName   string `json:"friendNickName"`
	FriendGender     int    `json:"friendGender"`
	FriendHeadImgUrl string `json:"friendHeadImgUrl"`

	QId      int    `json:"qid"`
	QText    string `json:"qtext"`
	QIconUrl string `json:"qiconUrl"`

	VoteFromUin        int64  `json:"voteFromUin"`
	VoteFromGender     int    `json:"voteFromGender"`
	VoteFromSchoolId   int    `json:"voteFromSchoolId"`
	VoteFromSchoolType int    `json:"voteFromSchoolType"`
	VoteFromSchoolName string `json:"voteFromSchoolName"`
	VoteFromGrade      int    `json:"voteFromGrade"`

	Ts int64 `json:"ts"`
}

func (this *FeedInfo) String() string {

	return fmt.Sprintf(`FeedInfo{VoteRecordId:%d FriendUin:%d FriendNickName:%s FriendGender:%d VoteFromUin:%d VoteFromGender:%d VoteFromSchoolName:%s  QId:%d, QText:%s}`,
		this.VoteRecordId, this.FriendUin, this.FriendNickName, this.FriendGender, this.VoteFromUin, this.VoteFromGender, this.VoteFromSchoolName, this.QId, this.QText)
}
