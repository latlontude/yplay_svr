package chengyuan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GetAccessTokenRsp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
}

func ExpressHandel(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start Express r:%+v", r)
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

		//if _, ok := codeOpenIdMap[code]; ok {
		appid := "wxc6a993bffd64bb9a"
		secret := "df1176a12952f19d0008facdab60edd6"

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
		codeOpenIdMap[code] = accessToken.OpenId

		if len(accessToken.OpenId) > 0 {
			ck1 := http.Cookie{Name: "openid", Value: fmt.Sprintf("%s", accessToken.OpenId), Path: "/"}
			http.SetCookie(w, &ck1)
		}
		//}

		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")

		htmlPath := "../download/sender/register.html"

		http.ServeFile(w, r, htmlPath)
	}
}

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start ImageHandler r:%+v", r)
	imagePath := "../download/" + r.URL.Path[1:]
	log.Debugf("imagePath:%s", imagePath)
	http.ServeFile(w, r, imagePath)
	log.Debugf("end ImageHandler")
	return
}

func JsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start JsHandler r:%+v", r)
	jsPath := "../download/" + r.URL.Path[1:]
	log.Debugf("JsPath:%s", jsPath)
	http.ServeFile(w, r, jsPath)
	log.Debugf("end jsHandler")
	return
}

func CssHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start CssHandler r:%+v", r)
	cssPath := "../download/" + r.URL.Path[1:]
	log.Debugf("cssPath:%s", cssPath)
	http.ServeFile(w, r, cssPath)
	log.Debugf("end CssHandler")
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

func CourierHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start CourierHandler r:%+v", r)

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")

	htmlPath := "../download/" + r.URL.Path[1:]
	log.Debugf("htmlPath:%s", htmlPath)
	http.ServeFile(w, r, htmlPath)
	log.Debugf("end CourierHandler")
	return
}

func IconsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start IconsHandler r:%+v", r)

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")

	htmlPath := "../download/" + r.URL.Path[1:]
	log.Debugf("htmlPath:%s", htmlPath)
	http.ServeFile(w, r, htmlPath)
	log.Debugf("end IconsHandler")
	return
}
func OwnerHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start OwnerHandler r:%+v", r)

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")

	htmlPath := "../download/" + r.URL.Path[1:]
	log.Debugf("htmlPath:%s", htmlPath)
	http.ServeFile(w, r, htmlPath)
	log.Debugf("end OwnerHandler")
	return
}
func SenderHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start SenderHandler r:%+v", r)

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")

	htmlPath := "../download/" + r.URL.Path[1:]
	log.Debugf("htmlPath:%s", htmlPath)
	http.ServeFile(w, r, htmlPath)
	log.Debugf("end SenderHandler")
	return
}

func PathHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start SenderHandler r:%+v", r)

	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")

	htmlPath := "../download/" + r.URL.Path[1:]
	log.Debugf("htmlPath:%s", htmlPath)
	http.ServeFile(w, r, htmlPath)
	log.Debugf("end SenderHandler")
	return
}
