package express

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type LoginReq struct {
	Code  string `schema:"code"`
	State string `schema:"state"`
}

type LoginRsp struct {
	UserInfo *UserInfo `json:"userInfo"`
}

type GetAccessTokenRsp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
}

type UserInfo struct {
	OpenId     string   `json:"openid"`
	NickName   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgUrl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionId    string   `json:"unionid"`
}

func doGetCode(req *LoginReq, r *http.Request) (rsp *LoginRsp, err error) {

	log.Debugf("r %+v\n req:%+v", r, req)

	appid := "wxc6a993bffd64bb9a"
	secret := "df1176a12952f19d0008facdab60edd6"
	//第一步：用户同意授权，获取code
	//code := r.Form.Get("code")
	code := req.Code
	//第二步：通过code换取网页授权access_token
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appid, secret, code)
	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	var accessToken GetAccessTokenRsp
	err = json.Unmarshal(body, &accessToken)
	if err != nil {

	}
	log.Debugf("accessToken :%+v", accessToken)

	//
	uinfo, err := GetUserInfo(accessToken.OpenId, accessToken.AccessToken)

	rsp = &LoginRsp{uinfo}

	return
}

func GetUserInfo(openid string, accessToken string) (userInfo *UserInfo, err error) {
	//第二步：通过code换取网页授权access_token
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		accessToken, openid)

	resp, err := http.Get(url)
	if err != nil {
		log.Debugf("resp:%+v \n err:%+v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("err:%+v", err)
	}

	userInfo = &UserInfo{}
	err = json.Unmarshal(body, userInfo)
	if err != nil {
		log.Debugf("err:%+v", err)
	}
	log.Debugf("userInfo :%+v", userInfo)

	return
}
