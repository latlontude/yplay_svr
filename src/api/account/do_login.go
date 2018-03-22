package account

import (
	"api/im"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"common/sms"
	"common/token"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"svr/cache"
	"svr/st"
	"time"
)

type LoginReq struct {
	Phone string `schema:"phone"`
	Code  string `schema:"code"`
	Uuid  int64  `schema:"uuid"`
}

type LoginRsp struct {
	Uin       int64  `json:"uin"`
	Token     string `json:"token"`
	Ver       int    `json:"ver"`
	IsNewUser int    `json:"isNewUser"`

	Info *st.UserProfileInfo `json:"info"`
}

func doLogin(req *LoginReq, r *http.Request) (rsp *LoginRsp, err error) {

	log.Debugf("phone %s, LoginReq %+v", req.Phone, req)

	rsp, err = Login(req.Phone, req.Code, req.Uuid)
	if err != nil {
		log.Errorf("phone %s, LoginRsp error, %s", req.Phone, err.Error())
		return
	}

	log.Debugf("phone %s, LoginRsp succ, %+v", req.Phone, rsp)

	return
}

func Login(phone string, code string, uuid int64) (rsp *LoginRsp, err error) {

	//2017-11-01 17:02
	if len(phone) == 0 || len(code) == 0 || uuid < constant.ENUM_DEVICE_UUID_MIN {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	if !sms.IsValidPhone(phone) {
		err = rest.NewAPIError(constant.E_INVALID_PHONE, "phone number invalid")
		log.Error(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SMS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%s", phone)
	valStr, err := app.Get(keyStr) //code 缓存63秒
	if err != nil {
		err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "code not exist in redis")
		log.Error(err.Error())
		return
	}

	//该手机号码不校验验证码
	if phone == "18682235582" {
		if len(code) != 4 {
			err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "invalid code")
			log.Error(err.Error())
			return
		}

	} else if code != valStr {
		err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "invalid code")
		log.Error(err.Error())
		return
	}

	uin, isNewUser, err := GetUinByPhone(phone)
	if err != nil {
		log.Error(err.Error())
		return
	}

	ttl := env.Config.Token.TTL
	token, err := token.GeneToken(uin, env.Config.Token.VER, ttl, uuid, "", "", "")
	if err != nil {
		log.Error(err.Error())
		return
	}

	app, err = myredis.GetApp(constant.ENUM_REDIS_APP_TOKEN)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = app.SetEx(fmt.Sprintf("%d", uin), token, uint32(ttl))
	if err != nil {
		log.Error(err.Error())
		return
	}

	info, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	rsp = &LoginRsp{uin, token, env.Config.Token.VER, isNewUser, info}

	return
}

func GetUinByPhone(phone string) (uin int64, isNewUser int, err error) {

	if phone != env.Config.Service.Phone {
		if !sms.IsValidPhone(phone) {
			err = rest.NewAPIError(constant.E_INVALID_PHONE, "phone number invalid")
			log.Error(err.Error())
			return
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	q := fmt.Sprintf(`select uin from profiles where phone = "%s"`, phone)

	rows, err := inst.Query(q)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	isNewUser = 1

	for rows.Next() {
		rows.Scan(&uin)
		isNewUser = 0
	}

	if isNewUser == 0 {
		return
	}

	stmt, err := inst.Prepare(`insert into profiles(uin, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType, schoolName, deptId, deptName, country, province, city, status, ts) 
		values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	userName := GeneUserName()
	nickName := ""
	headImgUrl := ""
	gender := 0
	age := 0
	grade := 0
	schoolId := 0
	schoolType := 0
	schoolName := ""
	country := ""
	province := ""
	city := ""
	status := 0
	deptId := 0
	deptName := ""

	ts := time.Now().Unix()

	res, err := stmt.Exec(0, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType, schoolName, deptId, deptName, country, province, city, status, ts)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	uin, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	go cache.UpdatePhone2Uin(phone, uin) //新注册用户更新缓存信息
	//go UpdateAddrBookInfo(phone, uin)  //等资料注册完整后,再更新通讯录里面的信息

	//到IM系统注册账号
	go im.SyncAccount(uin, fmt.Sprintf("%d", uin), "")

	return
}

func GeneUserName() (userName string) {

	ts := time.Now().UnixNano()

	b := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(b, ts)

	userName = base64.StdEncoding.EncodeToString(b)

	userName = strings.Replace(userName, "+", "Yz", -1)
	userName = strings.Replace(userName, "/", "Zy", -1)
	userName = strings.Replace(userName, "=", "", -1)

	return
}
