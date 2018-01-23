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

	log.Debugf("DownloadRedirectReq %+v", req)

	h := r.Header

	agentStr := ""
	if v, ok := h["User-Agent"]; ok {
		agentStr = strings.Join(v, ";")
	}

	redirectUrl := ""

	if REG_IOS.MatchString(agentStr) {

		redirectUrl = "https://itunes.apple.com/us/app/%E5%BC%80%E6%B5%AA/id1324604165?l=zh&ls=1&mt=8"

	} else if REG_MAC.MatchString(agentStr) {

		redirectUrl = "https://itunes.apple.com/us/app/%E5%BC%80%E6%B5%AA/id1324604165?l=zh&ls=1&mt=8"

	} else if REG_MI.MatchString(agentStr) {

		redirectUrl = "http://app.mi.com/details?id=com.zhihu.android"

	} else if REG_REDMI.MatchString(agentStr) {

		redirectUrl = "http://app.mi.com/details?id=com.zhihu.android"

	} else if REG_HUAWEI.MatchString(agentStr) {

		redirectUrl = "http://a.vmall.com/app/C10047082"

	} else if REG_HONOR.MatchString(agentStr) {

		redirectUrl = "http://a.vmall.com/app/C10047082"

	} else if REG_OPPO.MatchString(agentStr) {

		redirectUrl = "https://play.google.com/store/apps/details?id=com.zhihu.android"

	} else if REG_VIVO.MatchString(agentStr) {

	} else if REG_SM.MatchString(agentStr) {

		redirectUrl = "https://play.google.com/store/apps/details?id=com.zhihu.android"

	} else if REG_AND.MatchString(agentStr) {

		redirectUrl = "https://play.google.com/store/apps/details?id=com.zhihu.android"

	}

	rsp = &rest.RedirectInfo{redirectUrl}

	log.Debugf("RedirectInfo %+v", rsp)

	return
}
