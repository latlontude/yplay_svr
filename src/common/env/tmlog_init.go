package env

import (
	"common/tmlog"
	"strconv"
	"strings"
	"time"
)

// type tmlogConfigType struct {
// 	DebugPath  string `desc:"Debug路径"`
// 	TracePath  string `desc:"Trace路径"`
// 	InfoPath   string `desc:"Info路径"`
// 	ErrorPath  string `desc:"Error路径"`
// 	FatalPath  string `desc:"Fatal路径"`
// 	CronTime   string `desc:"日志文件切割周期(day,hour,ten)"`
// 	BufferSize int    `desc:"队列buffer长度"`
// 	FlushTimer int    `desc:"刷盘时间间隔(毫秒)"`
// 	DebugOpen  int    `desc:"是否启用终端调试"`
// 	LogLevels  string `desc:"日志级别"`
// }

// var tmlogConfig = &tmlogConfigType{
// 	DebugPath:  "../log/svr.log",
// 	TracePath:  "../log/svr.log",
// 	InfoPath:   "../log/svr.log",
// 	ErrorPath:  "../log/svr.log",
// 	FatalPath:  "../log/svr.log",
// 	CronTime:   "day",
// 	BufferSize: 10240,
// 	FlushTimer: 500,
// 	DebugOpen:  0,
// 	LogLevels:  "fatal,error,info,debug",
// }

func InitLog(logPath string, logFileName, logLevel string) error {

	config := make(map[string]string)

	path := logPath + "/" + logFileName

	config["log_notice_file_path"] = path
	config["log_debug_file_path"] = path
	config["log_trace_file_path"] = path
	config["log_fatal_file_path"] = path
	config["log_warning_file_path"] = path
	config["log_cron_time"] = "day"
	config["log_chan_buff_size"] = strconv.Itoa(4096000)
	config["log_flush_timer"] = strconv.Itoa(500)
	config["log_debug_open"] = strconv.Itoa(0)

	loglevel := 3 //minimal fatal and error

	for _, s := range strings.Split(logLevel, ",") {
		//for _, s := range strings.Split("fatal,error,info,debug", ",") {
		//输出日志的级别 (fatal:1,warngin:2,notice:4,trace:8,debug:16)
		switch s {
		case "fatal":
			loglevel |= 1
		case "warning", "error":
			loglevel |= 2
		case "notice", "info":
			loglevel |= 4
		case "trace":
			loglevel |= 8
		case "debug":
			loglevel |= 16
		}
	}
	config["log_level"] = strconv.Itoa(loglevel)

	// 启动 tmlog 工作协程, 可以理解为tmlog的服务器端
	go tmlog.Log_Run(config)

	//保证日志后台已经启动,否则容易出现空指针,导致panic
	time.Sleep(1 * time.Second)

	return nil
}

////////////////
type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

func NewLogger(id string) Logger {
	return tmlog.NewLogger(id)
}
