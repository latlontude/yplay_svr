package ddactivity

import (
	//	"common/auth"
	"common/env"
	//	"common/httputil"
	"common/constant"
	"common/myredis"
	"common/rest"
	"common/token"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type DDActivitySvrConfig struct {
	HttpServer struct {
		BindAddr string
	}

	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts    map[string]env.DataBase
	RedisInsts map[string]env.RedisInst

	RedisApps map[string]env.RedisApp

	Auth struct {
		Open            int //是否开启登录态验证 token算法验证和存储验证
		CheckTokenStore int //在开启验证的情况下, 是否开启存储校验
	}

	Token struct {
		TTL int //有效期
		VER int //版本
	}

	Activity struct {
		Schools string
	}

	BonusPool struct {
		Money int
	}

	NormalCall struct {
		Cnt int
	}
}

var (
	log         = env.NewLogger("ddactivity")
	OpenSchools map[int]int
	Config      DDActivitySvrConfig
)

func Init() (err error) {

	OpenSchools = make(map[int]int)

	schools := strings.Split(Config.Activity.Schools, ",")

	for _, s := range schools {
		sid, _ := strconv.Atoi(s)
		if sid > 0 {
			OpenSchools[sid] = 1
		}
	}

	log.Debugf("in ddactivity schools:%+v", OpenSchools)
	return
}

func MyHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start MyHandler r:%+v", r)
	path := strings.Trim(r.URL.Path, "/")
	log.Debugf("path:%s", path)

	file := "../download/" + path

	_, err := os.Stat(file)
	if err == nil {
		http.ServeFile(w, r, file)
	} else {
		log.Errorf("file :%s not exist!", file)
	}

	log.Debugf("end MyHandler")
	return
}

func WxLoadPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start WxLoadPageHandler r:%+v", r)
	r.ParseForm()
	if r.Method == "GET" {
		code := ""
		state := ""
		if _, ok := r.Form["code"]; ok {
			code = r.Form["code"][0]
		}
		if _, ok := r.Form["state"]; ok {
			state = r.Form["state"][0]
		}

		if len(code) == 0 || len(state) == 0 {
			log.Errorf("code or state is nil")
			return
		}
		openId, err := LoadPage(code, state)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		ck1 := http.Cookie{Name: "openId", Value: fmt.Sprintf("%s", openId), Path: "/"}
		http.SetCookie(w, &ck1)

		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")

		htmlPath := "../download/html/entry.html"
		http.ServeFile(w, r, htmlPath)

	}

	log.Debugf("end WxLoadPageHandler")
	return
}

func AppLoadPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start AppLoadPageHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		log.Debugf("token:%s", token)

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		shareSwitchStr := r.PostFormValue("shareswitch")
		if len(shareSwitchStr) == 0 {
			log.Errorf("no shareswitch param")
		} else if len(shareSwitchStr) > 0 && shareSwitchStr != "1" {
			log.Errorf("err shareswitch :%s", shareSwitchStr)
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		ck1 := http.Cookie{Name: "uin", Value: uinStr, Path: "/"}
		ck2 := http.Cookie{Name: "token", Value: token, Path: "/"}
		ck3 := http.Cookie{Name: "ver", Value: verStr, Path: "/"}
		ck4 := http.Cookie{Name: "shareswitch", Value: "1", Path: "/"}

		http.SetCookie(w, &ck1)
		http.SetCookie(w, &ck2)
		http.SetCookie(w, &ck3)
		http.SetCookie(w, &ck4)

		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")

		htmlPath := "../download/html/entry.html"
		http.ServeFile(w, r, htmlPath)
	} else if r.Method == "GET" {
		r.ParseForm()

		uinStr := ""
		token := ""
		verStr := ""
		shareSwitchStr := ""

		log.Debugf("r.Form:%+v", r.Form)

		if _, ok := r.Form["uin"]; ok {
			uinStr = r.Form["uin"][0]
		}
		if _, ok := r.Form["token"]; ok {
			token = r.Form["token"][0]
		}
		if _, ok := r.Form["ver"]; ok {
			verStr = r.Form["ver"][0]
		}

		if _, ok := r.Form["shareswitch"]; ok {
			shareSwitchStr = r.Form["shareswitch"][0]
		}

		if len(shareSwitchStr) > 0 && shareSwitchStr != "1" {
			log.Errorf("error shareswitch :%s", shareSwitchStr)
			return
		}

		if len(uinStr) == 0 || len(token) == 0 || len(verStr) == 0 {
			log.Errorf("no uin or token or ver param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		ck1 := http.Cookie{Name: "uin", Value: uinStr, Path: "/"}
		ck2 := http.Cookie{Name: "token", Value: token, Path: "/"}
		ck3 := http.Cookie{Name: "ver", Value: verStr, Path: "/"}
		ck4 := http.Cookie{Name: "shareswitch", Value: "1", Path: "/"}

		http.SetCookie(w, &ck1)
		http.SetCookie(w, &ck2)
		http.SetCookie(w, &ck3)
		http.SetCookie(w, &ck4)

		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")

		htmlPath := "../download/html/entry.html"
		http.ServeFile(w, r, htmlPath)

	}

	log.Debugf("end AppLoadPageHandler")
	return
}

func GetSingersFromPupuHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetSingersFromPupuHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}
		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &GetSingerFromPupuReq{uin, token, int(ver)}
		rsp, err := doGetSingersFromPupu(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetSingersFromPupuHandler")
	return
}

func GetVoteStatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetVoteStatusHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}
		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		type GetVoteStatusRsp struct {
			VoteStatus int `json:"voteStatus"`
		}

		status, err := Ishasvote(uin)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		var rsp GetVoteStatusRsp
		rsp.VoteStatus = status

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetVoteStatusHandler")
	return
}

func GetSingersFromWxHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetSingersFromWxHandler r:%+v", r)
	if r.Method == "POST" {
		openId := r.PostFormValue("openId")

		if len(openId) == 0 {
			log.Errorf("no openId param")
			return
		}

		pass, err := checkOpenId(openId)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("openId auth failed")
			return
		}

		req := &GetSingerFromWxReq{openId}
		rsp, err := doGetSingersFromWx(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetSingersFromWxHandler")
	return
}

func BeSingerFansFromPupuHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start BeSingerFansFromPupuHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		singerIdStr := r.PostFormValue("singerId")
		if len(singerIdStr) == 0 {
			log.Errorf("no singerId param")
			return
		}

		singerId, err := strconv.ParseInt(singerIdStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		req := &BeSingerFansFromPupuReq{uin, token, int(ver), int(singerId)}
		rsp, err := doBeSingerFansFromPupu(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end BeSingerFansFromPupuHandler")
	return
}

func BeSingerFansFromWxHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start BeSingerFansFromWxHandler r:%+v", r)
	if r.Method == "POST" {
		openId := r.PostFormValue("openId")
		if len(openId) == 0 {
			log.Errorf("no openId param")
			return
		}

		pass, err := checkOpenId(openId)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("openId auth failed")
			return
		}

		singerIdStr := r.PostFormValue("singerId")
		if len(singerIdStr) == 0 {
			log.Errorf("no singerId param")
			return
		}

		singerId, err := strconv.ParseInt(singerIdStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		req := &BeSingerFansFromWxReq{openId, int(singerId)}
		rsp, err := doBeSingerFansFromWx(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end BeSingerFansFromWxHandler")
	return
}

func GetSingersRankingListFromPupuHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetSingersRankingListFromPupuHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &GetSingerWithVoteFromPupuReq{uin, token, int(ver)}
		rsp, err := doGetSingersRankingListFromPupu(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetSingersRankingListFromPupuHandler")
	return
}

func GetSingersRankingListFromWxHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetSingersRankingListFromWxHandler r:%+v", r)
	if r.Method == "POST" {
		openId := r.PostFormValue("openId")
		if len(openId) == 0 {
			log.Errorf("no openId param")
			return
		}

		pass, err := checkOpenId(openId)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("openId auth failed")
			return
		}

		req := &GetSingerWithVoteFromWxReq{openId}
		rsp, err := doGetSingersRankingListFromWx(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetSingersRankingListFromWxHandler")
	return
}

func GetMySingerRankFromPupuHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetMySingerRankFromPupuHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &GetMySingerRankFromPupuReq{uin, token, int(ver)}
		rsp, err := doGetMySingerRankFromPupu(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetMySingerRankFromPupuHandler")
	return
}

func GetSingerDetailInfoFromPupuHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetSingerDetailInfoFromPupuHandler r:%+v", r)
	if r.Method == "POST" {
		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &GetMySingerRankFromPupuReq{uin, token, int(ver)}
		rsp, err := doGetMySingerRankFromPupu(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetSingerDetailInfoFromPupuHandler")
	return
}

func CallForSingerHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start CallForSingerHandler r:%+v", r)
	if r.Method == "POST" {

		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		typStr := r.PostFormValue("type")
		if len(typStr) == 0 {
			log.Errorf("no type param")
			return
		}

		typ, err := strconv.ParseInt(typStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		if typ < 1 || typ > 4 {
			log.Errorf("wrong call type :%d", typ)
			return
		}

		req := &CallForSingerReq{uin, token, int(ver), int(typ)}
		rsp, err := doCallForSinger(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end CallForSingerHandler")
	return
}

func GetCallTypeInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetCallTypeInfoHandler r:%+v", r)
	if r.Method == "POST" {

		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &GetCallTypeInfoReq{uin, token, int(ver)}
		rsp, err := doGetCallTypeInfo(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetCallTypeInfoHandler")
	return
}

func SingerRegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start SingerRegisterHandler r:%+v", r)
	if r.Method == "POST" {

		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		personalName := r.PostFormValue("personalName")
		if len(personalName) == 0 {
			log.Errorf("no personalName")
			return
		}

		activeHeadImgUrl := r.PostFormValue("activeHeadImgUrl")
		if len(activeHeadImgUrl) == 0 {
			log.Errorf("no activeHeadImgUrl")
			return
		}

		rankActiveHeadImgUrl := r.PostFormValue("rankActiveHeadImgUrl")
		if len(rankActiveHeadImgUrl) == 0 {
			log.Errorf("no rankActiveHeadImgUrl")
			return
		}

		singerDetailInfoImgUrl := r.PostFormValue("singerDetailInfoImgUrl")
		if len(singerDetailInfoImgUrl) == 0 {
			log.Errorf("no singerDetailInfoImgUrl")
			return
		}

		deptName := r.PostFormValue("deptName")
		if len(deptName) == 0 {
			log.Errorf("no deptName")
			return
		}

		declaration := r.PostFormValue("declaration")
		if len(declaration) == 0 {
			log.Errorf("no declaration")
			return
		}

		req := &SingerRegisterReq{uin, token, int(ver), personalName, activeHeadImgUrl, rankActiveHeadImgUrl, singerDetailInfoImgUrl, deptName, declaration}
		rsp, err := doSingerRegister(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end SingerRegisterHandler")
	return
}

func NormalCallForSingerHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start NormalCallForSingerHandler r:%+v", r)
	if r.Method == "POST" {

		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &NormalCallForSingerReq{uin, token, int(ver)}
		rsp, err := doNormalCallForSinger(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end NormalCallForSingerHandler")
	return
}

func GetCallInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetCallInfoHandler r:%+v", r)
	if r.Method == "POST" {

		uinStr := r.PostFormValue("uin")
		if len(uinStr) == 0 {
			log.Errorf("no uin param")
			return
		}

		uin, err := strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		token := r.PostFormValue("token")
		if len(token) == 0 {
			log.Errorf("no token param")
			return
		}

		verStr := r.PostFormValue("ver")
		if len(verStr) == 0 {
			log.Errorf("no ver param")
			return
		}

		ver, err := strconv.ParseInt(verStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pass, err := checkBaseParams(uin, token, int(ver))
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if !pass {
			log.Errorf("auth failed")
			return
		}

		req := &GetCallInfoReq{uin, token, int(ver)}
		rsp, err := doGetCallInfo(req, nil)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(&rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetCallInfoHandler")
	return
}

func HtmlHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start HtmlHandler r:%+v", r)

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")

	htmlPath := "../download/" + r.URL.Path[1:]
	log.Debugf("htmlPath:%s", htmlPath)
	http.ServeFile(w, r, htmlPath)
	log.Debugf("end HtmlHandler")
	return
}

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start ImageHandler r:%+v", r)
	/*
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	*/
	imagePath := "../download/" + r.URL.Path[1:]
	log.Debugf("imagePath:%s", imagePath)
	http.ServeFile(w, r, imagePath)
	log.Debugf("end ImageHandler")
	return
}

func JsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start JsHandler r:%+v", r)
	/*
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	*/
	jsPath := "../download/" + r.URL.Path[1:]
	log.Debugf("JsPath:%s", jsPath)
	http.ServeFile(w, r, jsPath)
	log.Debugf("end jsHandler")
	return
}

func CssHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start CssHandler r:%+v", r)
	/*
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	*/
	cssPath := "../download/" + r.URL.Path[1:]
	log.Debugf("cssPath:%s", cssPath)
	http.ServeFile(w, r, cssPath)
	log.Debugf("end CssHandler")
	return
}

func checkBaseParams(uin int64, tokenStr string, ver int) (pass bool, err error) {
	//不开启验证
	if Config.Auth.Open == 0 {
		/*	pass, err = uinExist(uin)
			if err != nil {
				log.Errorf(err.Error())
				pass = false
				return
			}
		*/
		pass = true
		return
	}

	log.Debugf("start check uin:%d, token:%s, ver:%d", uin, tokenStr, ver)

	t, err := token.DecryptToken(tokenStr, ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt fail")
		return
	}

	if t.Uin != int64(uin) || t.Ver != ver || t.Uuid < constant.ENUM_DEVICE_UUID_MIN {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "uin|ver|uuid invalid")
		return
	}

	ts := int(time.Now().Unix())
	if t.Ts < ts {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token expired")
		return
	}

	if Config.Auth.CheckTokenStore > 0 {

		app, err1 := myredis.GetApp(constant.ENUM_REDIS_APP_TOKEN)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_INVALID_SESSION, "redis app nil")
			return
		}

		tokenVal, err1 := app.Get(fmt.Sprintf("%d", uin))
		if err1 != nil {
			err = rest.NewAPIError(constant.E_INVALID_SESSION, "redis get token error")
			return
		}

		if tokenVal != tokenStr {
			err = rest.NewAPIError(constant.E_INVALID_SESSION, "token error")
			return
		}
	}

	pass = true
	log.Debugf("end check pass:%t", pass)
	return
}

func checkOpenId(openId string) (pass bool, err error) {
	//不开启验证
	if Config.Auth.Open == 0 {
		pass = true
		return
	}

	log.Debugf("start checkOpenId openId:%s", openId)

	p, err := userExist(openId)
	if err != nil {
		log.Errorf(err.Error())
		pass = false
		return
	} else {
		pass = p
	}

	log.Debugf("end checkOpenId pass:%t", pass)
	return
}
