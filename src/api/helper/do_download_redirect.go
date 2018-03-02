package helper

import (
	"common/rest"
	"net/http"
	"strings"
)

type DownloadRedirectReq struct {
	Sig string `schema:"sig"`
}

func doDownloadRedirect(req *DownloadRedirectReq, r *http.Request) (rsp *rest.RedirectInfo, err error) {

	h := r.Header

	agentStr := ""
	if v, ok := h["User-Agent"]; ok {
		agentStr = strings.Join(v, ";")
	}

	log.Debugf("DownloadRedirectReq %+v, agentStr %s", req, agentStr)

	redirectUrl := ""

	if REG_IOS.MatchString(agentStr) {

		redirectUrl = "https://itunes.apple.com/cn/app/%E5%99%97%E5%99%97/id1324604165?mt=8"

	} else if REG_MAC.MatchString(agentStr) {

		redirectUrl = "https://itunes.apple.com/cn/app/%E5%99%97%E5%99%97/id1324604165?mt=8"

	} else if REG_MI.MatchString(agentStr) {

		redirectUrl = "http://app.mi.com/detail/548878"

	} else if REG_MIX.MatchString(agentStr) {

		redirectUrl = "http://app.mi.com/detail/548878"

	} else if REG_REDMI.MatchString(agentStr) {

		redirectUrl = "http://app.mi.com/detail/548878"

	} else if REG_HUAWEI.MatchString(agentStr) {

		redirectUrl = "http://a.vmall.com/app/C100118923"

	} else if REG_HONOR.MatchString(agentStr) {

		redirectUrl = "http://a.vmall.com/app/C100118923"

	} else if REG_OPPO.MatchString(agentStr) {

		redirectUrl = "http://sj.qq.com/myapp/detail.htm?apkName=com.yeejay.yplay"

	} else if REG_VIVO.MatchString(agentStr) {

		redirectUrl = "http://info.appstore.vivo.com.cn/detail/2021675"

	} else if REG_SM.MatchString(agentStr) {

		redirectUrl = "http://sj.qq.com/myapp/detail.htm?apkName=com.yeejay.yplay"

	} else if REG_AND.MatchString(agentStr) {

		redirectUrl = "http://sj.qq.com/myapp/detail.htm?apkName=com.yeejay.yplay"

	}

	rsp = &rest.RedirectInfo{redirectUrl}

	log.Debugf("RedirectInfo %+v", rsp)

	return
}
