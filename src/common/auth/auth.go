package auth

import (
	"common/constant"
	"common/env"
	"common/myredis"
	"common/rest"
	"common/token"
	"fmt"
	"net/http"
	"time"
)

type t_AUTH int

const (
	AUTH = t_AUTH(0)
)

func (f t_AUTH) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	//不开启验证
	if env.Config.Auth.Open == 0 {
		return
	}

	var uin int
	var tokenStr string
	var ver int

	var err error

	if !rest.GetQueryParam(req, "uin", &uin) {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "uin param not exist")
		panic(err)
	}

	if !rest.GetQueryParam(req, "token", &tokenStr) {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token param not exist")
		panic(err)
	}

	if !rest.GetQueryParam(req, "ver", &ver) {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "ver param not exist")
		panic(err)
	}

	t, err := token.DecryptToken(tokenStr, ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt fail")
		panic(err)
	}

	if t.Uin != int64(uin) || t.Ver != ver || t.Uuid < constant.ENUM_DEVICE_UUID_MIN {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "uin|ver|uuid invalid")
		panic(err)
	}

	ts := int(time.Now().Unix())
	if t.Ts < ts {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token expired")
		panic(err)
	}

	//只进行算法校验 不进行存储校验
	if env.Config.Auth.CheckTokenStore > 0 {

		app, err := myredis.GetApp(constant.ENUM_REDIS_APP_TOKEN)
		if err != nil {
			err = rest.NewAPIError(constant.E_INVALID_SESSION, "redis app nil")
			panic(err)
		}

		tokenVal, err := app.Get(fmt.Sprintf("%d", uin))
		if err != nil {
			err = rest.NewAPIError(constant.E_INVALID_SESSION, "redis get token error")
			panic(err)
		}

		if tokenVal != tokenStr {
			err = rest.NewAPIError(constant.E_INVALID_SESSION, "token error")
			panic(err)
		}
	}

	if env.Config.Auth.CounterPV > 0 {
		go CounterUserPV(int64(uin))
	}

	return
}

func CounterUserPV(uin int64) (err error) {

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_PV_CNT)

	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "redis app nil")
		fmt.Printf(err.Error())
		return
	}

	y, m := time.Now().ISOWeek()

	_, err = app.Incr(fmt.Sprintf("%d_%d_%d", uin, y, m))
	if err != nil {
		fmt.Printf(err.Error())
		//return
	}

	//用户日活的统计,记录用户最近一次活跃的时间,每天在晚上12点之前统计当天的日活跃用户
	ts := time.Now().Unix()

	err = app.Set(fmt.Sprintf("%d_rts", uin), fmt.Sprintf("%d", ts))
	if err != nil {
		fmt.Printf(err.Error())
		//return
	}

	return
}

//无token校验
func Apify(fun interface{}) http.Handler {
	return rest.HandlerChain{
		rest.RPC(fun),
		rest.JSON,
	}
}

//token校验
func Apify2(fun interface{}) http.Handler {
	return rest.HandlerChain{
		AUTH,
		rest.RPC(fun),
		rest.JSON,
	}
}

//重定向
func Apify3(fun interface{}) http.Handler {
	return rest.HandlerChain{
		rest.RPC(fun),
		rest.REDIRECT,
	}
}

//无token校验
func Apify4(fun interface{}) http.Handler {
	return rest.HandlerChain{
		rest.RPC(fun),
		rest.RAW,
	}
}

//无token校验
func Apify5(fun interface{}) http.Handler {
	return rest.HandlerChain{
		rest.RPC(fun),
		rest.DOWNLOAD,
	}
}
