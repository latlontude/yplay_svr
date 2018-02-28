package account

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		//"/login":             auth.Apify(doLogin),              //登录
		"/sendsms":           auth.Apify(doSendSms),            //发送短信
		"/getnearestschools": auth.Apify2(doGetNearestSchools), //根据查询关键词获取学校列表
		"/getdeptsbyschool":  auth.Apify2(doGetDeptsBySchool),  //获取大学的学院信息

		"/searchschools": auth.Apify2(doSearchSchools), //根据查询关键词获取学校列表
		"/submitschool":  auth.Apify2(doSubmitSchool),  //提交增加学校信息

		"/decrypt2": auth.Apify2(doDecryptToken), //解析token信息

		"/logout":   auth.Apify(doLogout),   //登出
		"/cancell3": auth.Apify(doCancell3), //注销账号

		"/login2":          auth.Apify(doLogin2),          //登录 需要邀请码
		"/checkinvitecode": auth.Apify(doCheckInviteCode), //邀请码校验
	}

	log = env.NewLogger("account")
)
