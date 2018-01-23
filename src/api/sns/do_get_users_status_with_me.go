package sns

import (
	"common/constant"
	"common/rest"
	"encoding/json"
	"net/http"
	"svr/st"
	//"common/mydb"
	"fmt"
)

type GetUsersStatusWithMeReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Users string `schema:"users"`
}

type GetUsersStatusWithMeRsp struct {
	Infos []*UserStatusWithMeInfo `json:"infos"`
}

type UserStatusWithMeInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Status     int    `json:"status"`
}

func (this *UserStatusWithMeInfo) String() string {

	return fmt.Sprintf(`UserStatusWithMeInfo{Uin:%d, NickName:%d, HeadImgUrl:%s, Status:%d}`,
		this.Uin, this.NickName, this.HeadImgUrl, this.Status)
}

func doGetUsersStatusWithMe(req *GetUsersStatusWithMeReq, r *http.Request) (rsp *GetUsersStatusWithMeRsp, err error) {

	log.Debugf("GetUsersStatusWithMeReq %+v", req)

	sts, err := GetUsersStatusWithMe(req.Uin, req.Users)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	rsp = &GetUsersStatusWithMeRsp{sts}

	log.Debugf("GetUsersStatusWithMeRsp %+v", rsp)

	return
}

func GetUsersStatusWithMe(uin int64, users string) (sts []*UserStatusWithMeInfo, err error) {

	var userUins []int64

	err = json.Unmarshal([]byte(users), &userUins)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err.Error())
		return
	}

	sts = make([]*UserStatusWithMeInfo, 0)

	if len(userUins) == 0 {
		return
	}

	uinfoM, err := st.BatchGetUserProfileInfo(userUins)
	if err != nil {
		log.Error(err.Error())
		return
	}

	res, err := st.GetUinsStatusWithMe(uin, userUins)

	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, uid := range userUins {

		st := &UserStatusWithMeInfo{uid, "", "", 0}

		if val, ok := res[uid]; ok {
			st.Status = val
		}

		if ui, ok := uinfoM[uid]; ok {
			st.NickName = ui.NickName
			st.HeadImgUrl = ui.HeadImgUrl
		}

		sts = append(sts, st)
	}

	return
}
