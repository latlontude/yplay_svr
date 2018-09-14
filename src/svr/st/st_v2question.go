package st

import (
	"fmt"
)

type V2QuestionInfo struct {
	Qid           int                `json:"qid"`
	QTitle        string             `json:"qTitle"`
	QContent      string             `json:"qContent"`
	QImgUrls      string             `json:"qImgUrls"`
	OwnerInfo     *UserProfileInfo   `json:"ownerInfo"`
	QType         int                `json:"qType"`
	IsAnonymous   bool               `json:"isAnonymous"`
	CreateTs      int                `json:"createTs"`
	ModTs         int                `json:"modTs"`
	Longitude 	  float64 			  `schema:"longitude"`
	Latitude 	  float64 			  `schema:"latitude"`
	PoiTag 		  string 			  `schema:"poiTag"`
	Ext           string             `json:"ext"`
	AnswerCnt     int                `json:"answerCnt"`
	AccessCount   int                `json:"accessCount"`   //帖子访问数
	BestAnswer    *AnswersInfo       `json:"bestAnswer"`
	NewResponders []*UserProfileInfo `json:"newResponders"`
	Board         *BoardInfo		 `json:"board"`					//墙信息
}

func (this *V2QuestionInfo) String() string {

	return fmt.Sprintf(`V2QuestionInfo{Qid:%d QTitle:%s QContent:%s QImgUrls:%s OwnerInfo:%+v IsAnonymous:%t CreateTs:%d  ModTs:%d, Longitude:%f, Latitude:%f, poiTag:%s, AnswerCnt:%d}`,
		this.Qid, this.QTitle, this.QContent, this.QImgUrls, this.OwnerInfo, this.IsAnonymous, this.CreateTs, this.ModTs, this.Longitude, this.Latitude, this.PoiTag, this.AnswerCnt)
}
