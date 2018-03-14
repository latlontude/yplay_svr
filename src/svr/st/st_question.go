package st

import (
	"fmt"
)

type QuestionInfo struct {
	QId          int    `json:"qid"`
	QText        string `json:"qtext"`
	QIconUrl     string `json:"qiconUrl"`     //emoji
	OptionGender int    `json:"optionGender"` //选项性别要求 0无要求 1男性 2女性
	ReplyGender  int    `json:"replyGender"`  //答题者性别要求 0无要求 1男性 2女性
	SchoolType   int    `json:"schoolType"`   //schoolType 0通用 1初中 2高中 3大学
	DataSrc      int    `json:"dataSrc"`      //题目来源  0普通 1特制题库 2投稿 3运营
	Delivery     int    `json:"delivery"`     //投稿范围， 0 同校同年级可见，1 同校可见， 2 全网可见
	Status       int    `json:"status"`

	TagId       int    `json:"tagId"`
	TagName     string `json:"tagName"`
	SubTagId1   int    `json:"subTagId1"`
	SubTagName1 string `json:"subTagName1"`
	SubTagId2   int    `json:"subTagId2"`
	SubTagName2 string `json:"subTagName2"`
	SubTagId3   int    `json:"subTagId3"`
	SubTagName3 string `json:"subTagName3"`

	Ts int `json:"ts"`
}

type OptionInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	QId        int    `json:"qid"`
	BeSelCnt   int    `json:"beSelCnt"`
}

type OptionInfo2 struct {
	Uin      int64  `json:"uin"`
	NickName string `json:"nickName"`
	BeSelCnt int    `json:"beSelCnt"`
}

func (this *OptionInfo) String() string {

	return fmt.Sprintf(`OptionInfo{Uin:%d, NickName:%s, HeadImgUrl:%s, Gender:%d, QId:%d, BeSelCnt:%d}`,
		this.Uin, this.NickName, this.HeadImgUrl, this.Gender, this.QId, this.BeSelCnt)
}

func (this *OptionInfo2) String() string {

	return fmt.Sprintf(`OptionInfo2{Uin:%d, NickName:%s, BeSelCnt:%d}`,
		this.Uin, this.NickName, this.BeSelCnt)
}

func (this *QuestionInfo) String() string {

	return fmt.Sprintf(`QuestionInfo{QId:%d, QText:%s, QIconUrl:%s, OptionGender:%d, ReplyGender:%d, SchoolType:%d, dataSrc:%d,delivery:%d, tagId:%d, tagName:%s, subTagId1:%d,subTagName1:%s}`,
		this.QId, this.QText, this.QIconUrl, this.OptionGender, this.ReplyGender, this.SchoolType, this.DataSrc, this.Delivery, this.TagId, this.TagName, this.SubTagId1, this.SubTagName1)
}
