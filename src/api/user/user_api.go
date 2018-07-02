package user

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getuserprofile":            auth.Apify2(doGetUserProfile),           //拉取其他用户基础资料 不包括隐私资料
		"/getmyprofile":              auth.Apify2(doGetMyProfile),             //拉取用户自己基础资料 包含隐私资料
		"/getusergemstatinfo":        auth.Apify2(doGetUserGemStatInfo),       //拉取用户的钻石来源统计
		"/getusergemstatinfobycnt":   auth.Apify2(doGetUserGemStatInfoByCnt),  //拉取用户的钻石来源统计
		"/updateusergemstatinfo":     auth.Apify2(doUpdateUserGemStatInfo),    //更新用户的钻石统计信息
		"/getmyfriends":              auth.Apify2(doGetMyFriends),             //获取我的好友列表
		"/updateuserprofile":         auth.Apify2(doUpdateUserProfile),        //更新用户基础信息 头像/昵称/用户名称/性别
		"/updateschoolinfo":          auth.Apify2(doUpdateSchoolInfo),         //更新用户的学校信息
		"/schoolnameapprove":         auth.Apify(doApproveSchoolName),         //用户输入的学校名审核通过 方便运维处理 不加权限校验
		"/addschool":                 auth.Apify(doAddSchool),                 //在schools数据库表中增加新的学校,并将学校信息同步到mgo和es 中
		"/updateschoolfromdbtb":      auth.Apify(doUpdateSchool),              //将schools 数据库表中的学校信息同步到mgo和es 中
		"/reportuser":                auth.Apify(doReportUser),                //举报用户
		"/pullblackuser":             auth.Apify(doPullBlackUser),             //拉黑用户
		"/getmyblacklist":            auth.Apify(doGetMyBlacklist),            //获取我的黑名单用户
		"/removeuserfrommyblacklist": auth.Apify(doRemoveUserFromMyBlacklist), //从我的黑名单中移除用户

		"/getusersbyphone": auth.Apify2(doGetUsersbyPhone), //由手机号获取用户基本信息

		"/getmyprofilemodquota": auth.Apify2(doGetMyProfileModQuotaInfo), //拉取用户自己基础资料的修改配额（还能修改的次数）

		"/getheadimguploadsig": auth.Apify2(doGetHeadImgUploadSig), //拉取用户自己基础资料的修改配额（还能修改的次数）

		"/clearmods": auth.Apify(doClearMods), //清理历史修改的记录
		"/userextra": auth.Apify(doExUserInfo), //用户其他信息
	}

	log = env.NewLogger("user")
)
