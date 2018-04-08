package main

import (
	"common/env"
	"common/httputil"
	"common/mydb"
	"ddactivity"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var (
	confFile string
	log      = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/ddactivity_svr.conf", "默认配置文件路径")
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &ddactivity.Config))
	panicUnless(env.InitLog(ddactivity.Config.Log.LogPath, ddactivity.Config.Log.LogFileName, ddactivity.Config.Log.LogLevel))
	panicUnless(mydb.Init(ddactivity.Config.DbInsts))
	panicUnless(httputil.Init())

	http.HandleFunc("/wxvotepage", WxLoadPageHandler)
	http.HandleFunc("/appvotepage", AppLoadPageHandler)
	http.HandleFunc("/images/", ImageHandler)
	http.HandleFunc("/", MyHandler)
	http.ListenAndServe(ddactivity.Config.HttpServer.BindAddr, nil)
	log.Errorf("Starting ddactivity_svr...")

}

func MyHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.Trim(r.URL.Path, "/")
	if path == "MP_verify_cA6HNMxTCt2LwPpD.txt" {
		http.ServeFile(w, r, "../download/MP_verify_cA6HNMxTCt2LwPpD.txt")
	} else {
		io.WriteString(w, "welcome to pupu!\n")
	}
}

func WxLoadPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start WxLoadPageHandler r:%+v", r)
	r.ParseForm()
	if r.Method == "GET" {
		openId, err := ddactivity.LoadPage(r.Form["code"][0], r.Form["state"][0])
		if err != nil {
			io.WriteString(w, "welcome to pupu! \n")
		} else {
			ck1 := http.Cookie{Name: "openId", Value: fmt.Sprintf("%s", openId), Path: "/"}
			http.SetCookie(w, &ck1)

			htmlPath := "../download/index.html"
			http.ServeFile(w, r, htmlPath)
		}
	}

	log.Debugf("end WxLoadPageHandler")
}

func AppLoadPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start AppLoadPageHandler r:%+v", r)
	r.ParseForm()
	if r.Method == "GET" {
		ck1 := http.Cookie{Name: "uin", Value: r.Form["uin"][0], Path: "/"}
		ck2 := http.Cookie{Name: "token", Value: r.Form["token"][0], Path: "/"}
		ck3 := http.Cookie{Name: "ver", Value: r.Form["ver"][0], Path: "/"}

		http.SetCookie(w, &ck1)
		http.SetCookie(w, &ck2)
		http.SetCookie(w, &ck3)

		htmlPath := "../download/index.html"
		http.ServeFile(w, r, htmlPath)
	}

	log.Debugf("end AppLoadPageHandler")
}

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start ImageHandler r:%+v", r)
	imagePath := "../download/" + r.URL.Path[1:]
	log.Debugf("imagePath:%s", imagePath)
	http.ServeFile(w, r, imagePath)
	log.Debugf("end ImageHandler")
}
