package st

import (
	"fmt"
)

type BoardInfo struct {
	BoardId      int64  `json:"boardId"`
	BoardName    string `json:"boardName"`
	BoardIntro   string `json:"boardIntro"`
	BoardIconUrl string `json:"boardIconUrl"`
	BoardStatus  int    `json:"boardStatus"`

	SchoolId   int    `json:"schoolId"`
	SchoolName string `json:"schoolName"`
	SchoolType int    `json:"schoolType"`

	OwnerInfo *UserProfileInfo `json:"ownerInfo"`

	CreateTs int `json:"createTs"`
}

func (this *BoardInfo) String() string {

	return fmt.Sprintf(`BoardInfo{BoardId:%d BoardName:%s BoardIntro:%s BoardIconUrl:%s BoardStatus:%d SchoolId:%d SchoolName:%s  SchoolType:%d, CreateTs:%s, OwnerInfo:%+v}`,
		this.BoardId, this.BoardName, this.BoardIntro, this.BoardIconUrl, this.BoardStatus, this.SchoolId, this.SchoolName, this.SchoolType, this.CreateTs, this.OwnerInfo)
}
