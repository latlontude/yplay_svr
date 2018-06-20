package main

import (
	"common/env"
	"common/httputil"
	"common/mydb"
	"flag"
	"fmt"
	"net/http"
	"om"
	"os"
	"runtime"
)

var (
	confFile string
	log      = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/om_svr.conf", "默认配置文件路径")
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

	panicUnless(env.InitConfig(confFile, &om.Config))
	panicUnless(env.InitLog(om.Config.Log.LogPath, om.Config.Log.LogFileName, om.Config.Log.LogLevel))
	panicUnless(mydb.Init(om.Config.DbInsts))
	panicUnless(om.Init())
	panicUnless(httputil.Init())

	http.HandleFunc("/getallstory", om.GetAllStoryHandler)
	http.HandleFunc("/", om.Handler)

	log.Errorf("Starting om_svr...")

	http.ListenAndServe(om.Config.HttpServer.BindAddr, nil)
}
