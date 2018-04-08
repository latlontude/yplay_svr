package ddactivity

import (
	"common/auth"
	"common/env"
	"common/httputil"
	"strconv"
	"strings"
)

var (
	APIMap = httputil.APIMap{
		//"/loadpage": auth.Apify5(doLoadPage),
		"/singerregister":                auth.Apify2(doSingerRegister),                //歌手报名
		"/getsingersfrompupu":            auth.Apify2(doGetSingerFromPupu),             //pupu获取所有歌手信息
		"/getsingersfromwx":              auth.Apify2(doGetSingerFromWx),               //wx获取所有歌手信息
		"/besingerfansfrompupu":          auth.Apify2(doBeSingerFansFromPupu),          //噗噗注册用户成为歌手的粉丝
		"/besingerfansfromwx":            auth.Apify2(doBeSingerFansFromWx),            //微信用户成为歌手的粉丝
		"/getsingersrankinglistfrompupu": auth.Apify2(doGetSingersRankingListFromPupu), //pupu获取歌手得票信息
		"/getsingersrankinglistfromwx":   auth.Apify2(doGetSingersRankingListFromWx),   //wx获取歌手得票信息
		"/docallforsinger":               auth.Apify2(doCallForSinger),                 //为歌手打call
		"/getcalltypeinfo":               auth.Apify2(doGetCallTypeInfo),               //获取当前为爱豆打call进度
	}

	log         = env.NewLogger("ddactivity")
	OpenSchools map[int]int
)

type DDActivitySvrConfig struct {
	HttpServer struct {
		BindAddr string
	}

	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts  map[string]env.DataBase
	Activity struct {
		Schools string
	}
}

var (
	Config DDActivitySvrConfig
)

func Init() (err error) {

	OpenSchools = make(map[int]int)

	schools := strings.Split(Config.Activity.Schools, ",")

	for _, s := range schools {
		sid, _ := strconv.Atoi(s)
		if sid > 0 {
			OpenSchools[sid] = 1
		}
	}

	return
}
