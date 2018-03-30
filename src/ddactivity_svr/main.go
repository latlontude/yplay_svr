package main

import (
	"common/env"
	"common/httputil"
	"ddactivity"
	"flag"
	"fmt"
	"os"
	"runtime"
)

//计算2度好友关系的SVR配置
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

var (
	confFile string
	config   DDActivityConfig

	log = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/dd_activity.conf", "默认配置文件路径")
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

	panicUnless(env.InitConfig(confFile, &config))
	//panicUnless(mydb.Init(config.DbInsts))
	panicUnless(env.InitLog(config.Log.LogPath, config.Log.LogFileName, config.Log.LogLevel))
	//panicUnless(myredis.Init(config.RedisInsts, config.RedisApps))
	panicUnless(httputil.Init())

	httputil.HandleAPIMap("/dd/activity/", ddactivity.APIMap)

	log.Errorf("Starting ddactivity_svr...")
	panicUnless(httputil.ListenHttp(config.HttpServer.BindAddr))

}
