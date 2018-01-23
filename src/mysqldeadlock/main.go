package main

import (
	"common/constant"
	"common/env"
	"common/mydb"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	confFile string
	log      = env.NewLogger("deadlock")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/frozen_monitor_svr.conf", "默认配置文件路径")
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

	panicUnless(env.InitConfig(confFile, &env.FrozenMonitorConfig))
	panicUnless(env.InitLog(env.FrozenMonitorConfig.Log.LogPath, env.FrozenMonitorConfig.Log.LogFileName, env.FrozenMonitorConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.FrozenMonitorConfig.DbInsts))

	log.Errorf("start server....")

	go Monitor()
	go Monitor()

	for {
		select {
		case <-time.After(time.Second * time.Duration(10)):
			log.Errorf("new loop...")
		}
	}
}

func Monitor() (err error) {

	log.Errorf("begin monitor...")

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		log.Error(err.Error())
		return
	}

	now := time.Now().Unix()
	uin := 100032

	sql := fmt.Sprintf(`update freezingStatus set freezeTs = 0, ts = %d where uin = %d`, now, uin)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Errorf(err.Error())
	}

	log.Errorf("end monitor...")

	return
}
