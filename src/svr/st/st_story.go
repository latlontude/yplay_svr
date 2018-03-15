package st

import (
	"fmt"
)

type StoryInfo struct {
	StoryId         int64  `json:"storyId"`
	Type            int    `json:"type"`
	Text            string `json:"text"`
	Data            string `json:"data"`
	Uin             int64  `json:"uin"`
	ThumbnailImgUrl string `json:"thumbnailImgUrl"`
	ViewCnt         int    `json:"viewCnt"`
	Ts              int64  `json:"ts"`
}

type RetStoryInfo struct {
	StoryId         int64  `json:"storyId"`
	Type            int    `json:"type"`
	Text            string `json:"text"`
	Data            string `json:"data"`
	Uin             int64  `json:"uin"`
	NickName        string `json:"nickName"`
	HeadImgUrl      string `json:"headImgUrl"`
	ThumbnailImgUrl string `json:"thumbnailImgUrl"`
	ViewCnt         int    `json:"viewCnt"`
	Ts              int64  `json:"ts"`
}

func (this *StoryInfo) String() string {
	return fmt.Sprintf(`StoryInfo{StoryId:%d Type:%d Data:%s Uin:%d  Text:%s ThumbnailImgUrl:%s ViewCnt:%d Ts:%d}`,
		this.StoryId, this.Type, this.Data, this.Uin, this.Text, this.ThumbnailImgUrl, this.ViewCnt, this.Ts)
}
