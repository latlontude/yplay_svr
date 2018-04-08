package ddactivity

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	WXAPPID     = "wxc6a993bffd64bb9a"
	WXAPPSECRET = "df1176a12952f19d0008facdab60edd6"
)

type LoadPageReq struct {
	Code  string `schema:"code"`
	State string `schema:"state"` // pupu
}

type TokenRspData struct {
	Access_token  string `json:"access_token"`
	Expires_in    int    `json:"expires_in"`
	Refresh_token string `json:"refresh_token"`
	Openid        string `json:"openid"`
	Scope         string `json:"scope"`
}

type ErrRspData struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

type WxUserInfo struct {
	Openid     string   `json:"openid"`
	NickName   string   `json:"nickname"`
	Sex        int      `json:"sex"` // 1为男性，2为女性，0 未知
	Language   string   `json:"language"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgUrl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionId    string   `json:"unionid"`
}

func LoadPage(code, state string) (openId string, err error) {
	log.Debugf(" start LoadPage code:%s state:%s", code, state)
	accessToken, openid, err := getAccessToken(code)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	if len(accessToken) == 0 || len(openid) == 0 {
		log.Debugf("accessToken or openid is null")
		return
	}

	b, err := userExist(openid)
	if err != nil || !b {
		info, err1 := getWxUserInfo(accessToken, openid)
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}
		err1 = storeUserInfo(info)
		if err1 != nil {
			log.Errorf(err1.Error())
			return
		}
	}

	openId = openid
	log.Debugf("end LoadPage")
	return
}

func getAccessToken(code string) (accessToken, openid string, err error) {
	log.Debugf("start getAccessToken code:%s", code)

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", WXAPPID, WXAPPSECRET, code)
	ret, err := http.Get(url)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	b, err := ioutil.ReadAll(ret.Body)
	ret.Body.Close()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var data TokenRspData
	err = json.Unmarshal(b, &data)
	if err != nil {
		var errData ErrRspData
		err = json.Unmarshal(b, &errData)
		if err != nil {
			log.Debugf("errData:%+v", errData)
		}
	} else {
		log.Debugf("data:%+v", data)
		accessToken = data.Access_token
		openid = data.Openid
	}

	log.Debugf("end getAccessToken accessToken:%s", accessToken)
	return
}

func getWxUserInfo(access_token, openid string) (info WxUserInfo, err error) {
	log.Debugf("start getWxUserInfo access_token:%s openid:%s", access_token, openid)

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", access_token, openid)
	ret, err := http.Get(url)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	b, err := ioutil.ReadAll(ret.Body)
	ret.Body.Close()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var data WxUserInfo
	err = json.Unmarshal(b, &data)
	if err != nil {
		var errData ErrRspData
		err = json.Unmarshal(b, &errData)
		if err != nil {
			log.Debugf("errData:%+v", errData)
		}
	} else {
		log.Debugf("data:%+v", data)
		info = data
	}

	log.Debugf("end getWxUserInfo info:%+v", info)
	return
}

func storeUserInfo(user WxUserInfo) (err error) {
	log.Debugf("start storeUserInfo user:%+v", user)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	ts := time.Now().Unix()
	status := 0
	privilegeStr := ""
	for _, val := range user.Privilege {
		privilegeStr += fmt.Sprintf("%s,", val)
	}
	if len(privilegeStr) > 0 {
		privilegeStr = privilegeStr[:len(privilegeStr)-1]
	}

	stmt, err := inst.Prepare(`insert into wxUserInfo values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(0, user.Openid, user.NickName, user.Sex, user.Language, user.Province, user.City, user.Country, user.HeadImgUrl, privilegeStr, user.UnionId, ts, status)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	log.Debugf("end storeUserInfo")
	return
}

func userExist(openid string) (exist bool, err error) {
	log.Debugf("start userExist openid:%s", openid)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf("select openId from wxUserInfo where openId = '%s'", openid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	find := false
	for rows.Next() {
		find = true
		break
	}
	exist = find

	log.Debugf("end userExist exist:%t", exist)
	return
}
