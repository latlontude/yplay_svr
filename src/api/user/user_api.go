package user

import (
	"common/auth"
	"common/env"
	"common/httputil"
)

var (
	APIMap = httputil.APIMap{
		"/getuserprofile":        auth.Apify2(doGetUserProfile),        //拉取其他用户基础资料 不包括隐私资料
		"/getmyprofile":          auth.Apify2(doGetMyProfile),          //拉取用户自己基础资料 包含隐私资料
		"/getusergemstatinfo":    auth.Apify2(doGetUserGemStatInfo),    //拉取用户的钻石来源统计
		"/updateusergemstatinfo": auth.Apify2(doUpdateUserGemStatInfo), //更新用户的钻石统计信息
		"/getmyfriends":          auth.Apify2(doGetMyFriends),          //获取我的好友列表
		"/updateuserprofile":     auth.Apify2(doUpdateUserProfile),     //更新用户基础信息 头像/昵称/用户名称/性别
		"/updateschoolinfo":      auth.Apify2(doUpdateSchoolInfo),      //更新用户的学校信息
		"/schoolnameapprove":     auth.Apify(doApproveSchoolName),      //用户输入的学校名审核通过 方便运维处理 不加权限校验

		"/getusersbyphone": auth.Apify2(doGetUsersbyPhone), //由手机号获取用户基本信息

		"/getmyprofilemodquota": auth.Apify2(doGetMyProfileModQuotaInfo), //拉取用户自己基础资料的修改配额（还能修改的次数）

		"/getheadimguploadsig": auth.Apify2(doGetHeadImgUploadSig), //拉取用户自己基础资料的修改配额（还能修改的次数）

		"/clearmods": auth.Apify(doClearMods), //清理历史修改的记录
	}

	log = env.NewLogger("user")
)
