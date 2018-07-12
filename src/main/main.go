package main

import (
	"api/account"
	"api/activity"
	"api/addr"
	"api/answer"
	"api/board"
	"api/comment"
	"api/feed"
	"api/geneqids"
	"api/helper"
	"api/im"
	"api/like"
	"api/msg"
	"api/notify"
	"api/push"
	"api/question"
	"api/sns"
	"api/story"
	"api/submit"
	"api/user"
	"api/vote"
	"api/experience"
	"common/env"
	"common/httputil"
	"common/mydb"
	"common/mymgo"
	"common/myredis"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"svr/cache"
	"time"
)

var (
	confFile string
	log      = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/yplay_svr.conf", "默认配置文件路径")

	rand.Seed(time.Now().UnixNano())
}

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &env.Config))
	panicUnless(env.InitLog(env.Config.Log.LogPath, env.Config.Log.LogFileName, env.Config.Log.LogLevel))
	panicUnless(mydb.Init(env.Config.DbInsts))
	panicUnless(myredis.Init(env.Config.RedisInsts, env.Config.RedisApps))
	panicUnless(mymgo.Init(env.Config.MgoInsts))
	panicUnless(httputil.Init())

	panicUnless(cache.Init())
	panicUnless(im.Init()) //主要是好友动态的channnel routine创建
	panicUnless(activity.Init())

	httputil.HandleAPIMap("/api/account/", account.APIMap)
	httputil.HandleAPIMap("/api/answer/", answer.APIMap)
	httputil.HandleAPIMap("/api/board/", board.APIMap)
	httputil.HandleAPIMap("/api/like/", like.APIMap)
	httputil.HandleAPIMap("/api/comment/", comment.APIMap)
	httputil.HandleAPIMap("/api/question/", question.APIMap)
	httputil.HandleAPIMap("/api/feed/", feed.APIMap)
	httputil.HandleAPIMap("/api/msg/", msg.APIMap)
	httputil.HandleAPIMap("/api/sns/", sns.APIMap)
	httputil.HandleAPIMap("/api/user/", user.APIMap)
	httputil.HandleAPIMap("/api/vote/", vote.APIMap)
	httputil.HandleAPIMap("/api/addr/", addr.APIMap)
	httputil.HandleAPIMap("/api/im/", im.APIMap)
	httputil.HandleAPIMap("/api/push/", push.APIMap)
	httputil.HandleAPIMap("/api/notify/", notify.APIMap)
	httputil.HandleAPIMap("/api/submit/", submit.APIMap)
	httputil.HandleAPIMap("/api/geneqids/", geneqids.APIMap)
	httputil.HandleAPIMap("/api/helper/", helper.APIMap)
	httputil.HandleAPIMap("/api/story/", story.APIMap)
	httputil.HandleAPIMap("/svr/cache/", cache.APIMap)
	httputil.HandleAPIMap("/api/activity/", activity.APIMap)
	httputil.HandleAPIMap("/api/experience/", experience.APIMap)

	log.Errorf("Starting yplay_svr...")
	panicUnless(httputil.ListenHttp(env.Config.HttpServer.BindAddr))
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
