package st

import (
	"fmt"
)

type BoardInfo struct {
	BoardId      int    `json:"boardId"`
	BoardName    string `json:"boardName"`
	BoardIntro   string `json:"boardIntro"`
	BoardIconUrl string `json:"boardIconUrl"`
	BoardStatus  int    `json:"boardStatus"`

	SchoolId   int              `json:"schoolId"`
	SchoolName string           `json:"schoolName"`
	SchoolType int              `json:"schoolType"`
	Longitude  float64          `json:"longitude"`
	Latitude   float64          `json:"latitude"`
	OwnerInfo  *UserProfileInfo `json:"ownerInfo"`
	FollowCnt  int              `json:"followCnt"`
	CreateTs   int              `json:"createTs"`
	IsAdmin    bool             `json:"isAdmin"`
}

func (this *BoardInfo) String() string {

	return fmt.Sprintf(`BoardInfo{BoardId:%dBoardName:%s BoardIntro:%s BoardIconUrl:%s BoardStatus:%d SchoolId:%d 
SchoolName:%s  SchoolType:%d, longitude:%f,latitude:%f,CreateTs:%d, OwnerInfo:%+v, FollowCnt:%d, isAdmin:%v}`,
		this.BoardId, this.BoardName, this.BoardIntro, this.BoardIconUrl, this.BoardStatus,
		this.SchoolId, this.SchoolName, this.SchoolType, this.Longitude, this.Latitude, this.CreateTs, this.OwnerInfo, this.FollowCnt, this.IsAdmin)
}
