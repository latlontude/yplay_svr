package main

import (
	"common/env"
	"common/httputil"
	"common/mydb"
	"ddsinger"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var (
	confFile string
	log      = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/ddsinger_svr.conf", "默认配置文件路径")
	rand.Seed(time.Now().UnixNano())
}

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &env.WxPublicConfig))
	panicUnless(env.InitLog(env.WxPublicConfig.Log.LogPath, env.WxPublicConfig.Log.LogFileName, env.WxPublicConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.WxPublicConfig.DbInsts))
	panicUnless(httputil.Init())

	httputil.HandleAPIMap("/api/", ddsinger.APIMap)

	log.Errorf("Starting ddsinger_wxpublic_svr...")
	panicUnless(httputil.ListenHttp(env.WxPublicConfig.HttpServer.BindAddr))
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
