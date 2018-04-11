package main

import (
	"common/env"
	"common/httputil"
	"common/mydb"
	"common/myredis"
	"ddactivity"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
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
	panicUnless(myredis.Init(ddactivity.Config.RedisInsts, ddactivity.Config.RedisApps))
	panicUnless(ddactivity.Init())
	panicUnless(httputil.Init())

	http.HandleFunc("/votepage", ddactivity.WxLoadPageHandler)
	http.HandleFunc("/appvotepage", ddactivity.AppLoadPageHandler)
	http.HandleFunc("/getsingersfrompupu", ddactivity.GetSingersFromPupuHandler)
	http.HandleFunc("/getsingersfromwx", ddactivity.GetSingersFromWxHandler)
	http.HandleFunc("/besingerfansfrompupu", ddactivity.BeSingerFansFromPupuHandler)
	http.HandleFunc("/besingerfansfromwx", ddactivity.BeSingerFansFromWxHandler)
	http.HandleFunc("/getsingersrankinglistfrompupu", ddactivity.GetSingersRankingListFromPupuHandler)
	http.HandleFunc("/getsingersrankinglistfromwx", ddactivity.GetSingersRankingListFromWxHandler)
	http.HandleFunc("/docallforsinger", ddactivity.CallForSingerHandler)
	http.HandleFunc("/getcalltypeinfo", ddactivity.GetCallTypeInfoHandler)
	http.HandleFunc("/singerregister", ddactivity.SingerRegisterHandler)
	http.HandleFunc("/call", ddactivity.NormalCallForSingerHandler)
	http.HandleFunc("/getcallinfo", ddactivity.GetCallInfoHandler)
	http.HandleFunc("/getvotestatus", ddactivity.GetVoteStatusHandler)

	http.HandleFunc("/images/", ddactivity.ImageHandler)
	http.HandleFunc("/javascript/", ddactivity.JsHandler)
	http.HandleFunc("/styles/", ddactivity.CssHandler)
	http.HandleFunc("/html/", ddactivity.HtmlHandler)
	http.HandleFunc("/", ddactivity.MyHandler)

	log.Errorf("Starting ddactivity_svr...")

	http.ListenAndServe(ddactivity.Config.HttpServer.BindAddr, nil)
}
