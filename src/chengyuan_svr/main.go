package main

import (
	"chengyuan"
	"common/env"
	"common/httputil"
	"common/mydb"
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
	flag.StringVar(&confFile, "f", "../etc/chengyuan_svr.conf", "默认配置文件路径")
	rand.Seed(time.Now().UnixNano())
}

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &env.WxPublicConfig))
	panicUnless(env.InitLog(env.WxPublicConfig.Log.LogPath, env.WxPublicConfig.Log.LogFileName, env.WxPublicConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.WxPublicConfig.DbInsts))
	panicUnless(httputil.Init())

	httputil.HandleAPIMap("/api/", chengyuan.APIMap)

	log.Errorf("Starting chengyuan_wxpublic_svr...")
	panicUnless(httputil.ListenHttp(env.WxPublicConfig.HttpServer.BindAddr))
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
