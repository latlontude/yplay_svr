package env

import (
//"os"
)

var (
	Config              YPLAYSvrConfig
	Calc2DegreeConfig   Calc2DegreeSvrConfig
	FrozenMonitorConfig FrozenMonitorSvrConfig
	PreGeneQIdsConfig   PreGeneQIdsSvrConfig
	WxPublicConfig      WxPublicSvrConfig
)

type DataBase struct {
	Host        string
	Port        int
	User        string
	Passwd      string
	DbName      string
	MaxOpenConn int
	MaxIdleConn int
}

type RedisInst struct {
	Host        string
	Port        int
	MaxOpenConn int
	MaxIdleConn int
	Passwd      string
}

type MgoInst struct {
	Hosts  string
	User   string
	Passwd string
	DbName string

	MaxConn int
	TimeOut int
}

type RedisApp struct {
	InstName string
	AppId    string
}

type YPLAYSvrConfig struct {
	HttpServer struct {
		BindAddr string
	}

	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts    map[string]DataBase
	RedisInsts map[string]RedisInst
	MgoInsts   map[string]MgoInst

	RedisApps map[string]RedisApp

	Auth struct {
		Open            int //是否开启登录态验证 token算法验证和存储验证
		CheckTokenStore int //在开启验证的情况下, 是否开启存储校验
		CounterPV       int //是否统计用户的PV 按年+周
	}

	Token struct {
		TTL int //有效期
		VER int //版本
	}

	Sms struct {
		TTL              int
		InviteFriendSend int
	}

	Feed struct {
		TrimCnt          int
		MaxCnt           int
		PushGap          int //动态聚合发送push的间隔
		PushMergeUserCnt int //聚合的用户数
	}

	Story struct {
		TrimCnt int
		PushGap int //新闻聚合发送push的间隔
	}

	Sensitive struct {
		Set string //昵称敏感词库
	}

	Addr struct {
		UploadBatchSize int
	}

	Vote struct {
		FreezeDuration int
	}

	Profile struct {
		ModMaxCnt       int
		GenderModMaxCnt int
	}

	Service struct { //噗噗客服登录账号
		Phone string
		Code  string
	}

	InnerTest struct { //内部测试手机号
		Phones string
		Code   string
	}

	WhiteList struct {
		Phones string
	}

	Activity struct {
		Schools string
	}

	WeekRankBlacklist struct {
		Uins string
	}
}

//计算2度好友关系的SVR配置
type Calc2DegreeSvrConfig struct {
	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts    map[string]DataBase
	RedisInsts map[string]RedisInst
	RedisApps  map[string]RedisApp

	CalcGap int
}

//冷冻状态监控的服务发push
type FrozenMonitorSvrConfig struct {
	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts        map[string]DataBase
	MonitorDBGap   int
	MonitorTaskGap int
	FreezeDuration int
}

type PreGeneQIdsSvrConfig struct {
	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts    map[string]DataBase
	RedisInsts map[string]RedisInst
	RedisApps  map[string]RedisApp

	GeneModInfo struct {
		ModName string //include or exclude
		Ids     string //includ mod only gene ids, exclude mod gene alluins except ids
	}

	GeneGap         int
	ReStartGeneFlag int
}

type WxPublicSvrConfig struct {
	HttpServer struct {
		BindAddr string
	}

	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts map[string]DataBase
}
