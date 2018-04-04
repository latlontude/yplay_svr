package main

import (
	"common/env"
	"common/httputil"
	"common/mydb"
	"ddactivity"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

/*
type DDActivityConfig struct {
	HttpServer struct {
		BindAddr string
	}

	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}
}
*/

var (
	confFile string
	//config   DDActivityConfig

	log = env.NewLogger("main")
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

	panicUnless(env.InitConfig(confFile, &env.DdActivityConfig))
	panicUnless(env.InitLog(env.DdActivityConfig.Log.LogPath, env.DdActivityConfig.Log.LogFileName, env.DdActivityConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.DdActivityConfig.DbInsts))
	panicUnless(httputil.Init())

	http.HandleFunc("/votepage", LoadPageHandler)
	http.HandleFunc("/images/", ImageHandler)
	http.HandleFunc("/", MyHandler)
	http.ListenAndServe(":80", nil)
	log.Errorf("Starting ddactivity_svr...")

}

func MyHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.Trim(r.URL.Path, "/")
	if path == "MP_verify_cA6HNMxTCt2LwPpD.txt" {
		t, err := template.ParseFiles("../download/MP_verify_cA6HNMxTCt2LwPpD.txt")
		if err != nil {
			log.Errorf(err.Error())
		}
		t.Execute(w, nil)
	} else {
		io.WriteString(w, "welcome to pupu!\n")
	}
}

func LoadPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start LoadPageHandler r:%+v", r)
	r.ParseForm()
	if r.Method == "GET" {
		err := ddactivity.LoadPage(r.Form["code"][0], r.Form["state"][0])
		if err != nil {
			io.WriteString(w, "welcome to pupu! \n")
		} else {
			t, err := template.ParseFiles("../download/index.html")
			if err != nil {
				log.Errorf(err.Error())
			}
			t.Execute(w, nil)
		}
	}

	log.Debugf("end LoadPageHandler")
}

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start ImageHandler r:%+v", r)
	imagePath := "../download/" + r.URL.Path[1:]
	log.Debugf("imagePath:%s", imagePath)
	http.ServeFile(w, r, imagePath)
	log.Debugf("end ImageHandler")
}
