package st

import (
	"fmt"
)

type V2QuestionInfo struct {
	Qid         int              `json:"qid"`
	QTitle      string           `json:"qTitle"`
	QContent    string           `json:"qContent"`
	QImgUrls    string           `json:"qImgUrls"`
	OwnerInfo   *UserProfileInfo `json:"ownerInfo"`
	IsAnonymous bool             `json:"isAnonymous"`
	CreateTs    int              `json:"createTs"`
	ModTs       int              `json:"modTs"`
	AnswerCnt   int              `json:"answerCnt"`
	BestAnswer  *AnswersInfo     `json:"bestAnswer"`
}

func (this *V2QuestionInfo) String() string {

	return fmt.Sprintf(`V2QuestionInfo{Qid:%d QTitle:%s QContent:%s QImgUrls:%s OwnerInfo:%+v IsAnonymous:%t CreateTs:%d  ModTs:%d, AnswerCnt:%d}`,
		this.Qid, this.QTitle, this.QContent, this.QImgUrls, this.OwnerInfo, this.IsAnonymous, this.CreateTs, this.ModTs, this.AnswerCnt)
}
