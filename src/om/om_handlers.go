package om

import (
	"common/env"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type OmSvrConfig struct {
	HttpServer struct {
		BindAddr string
	}

	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	DbInsts map[string]env.DataBase

	Auth struct {
		Open            int //是否开启登录态验证 token算法验证和存储验证
		CheckTokenStore int //在开启验证的情况下, 是否开启存储校验
	}

	Token struct {
		TTL int //有效期
		VER int //版本
	}

	BlackList struct {
		Uins string
	}
}

var (
	log    = env.NewLogger("om")
	Config OmSvrConfig
)

const (
	DEFAULT_PAGE_SIZE = 10
)

func Init() (err error) {

	return
}

func GetAllStoryHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start GetAllStoryHandler r:%+v", r)
	if r.Method == "POST" {
		sessionId := r.PostFormValue("sessionId")
		if len(sessionId) == 0 {
			log.Errorf("no sessionId param")
			return
		}

		pageNumStr := r.PostFormValue("pageNum")
		if len(pageNumStr) == 0 {
			log.Errorf("no pageNum param")
			return
		}

		pageNum, err := strconv.ParseInt(pageNumStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		pageSizeStr := r.PostFormValue("pageSize")
		if len(pageSizeStr) == 0 {
			log.Errorf("no pageSize param")
			return
		}

		pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		uinStr := r.PostFormValue("uin")
		var uin int64

		if len(uinStr) != 0 {
			uid, err := strconv.ParseInt(uinStr, 10, 64)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			uin = uid
		}

		msgTypeStr := r.PostFormValue("msgType")
		var msgType int64

		if len(msgTypeStr) != 0 {
			typ, err := strconv.ParseInt(msgTypeStr, 10, 64)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			msgType = typ
		}

		tsStartStr := r.PostFormValue("tsStart")
		var tsStart int64

		if len(tsStartStr) != 0 {
			start, err := strconv.ParseInt(tsStartStr, 10, 64)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			tsStart = start
		}

		tsEndStr := r.PostFormValue("tsEnd")
		var tsEnd int64

		if len(tsEndStr) != 0 {
			end, err := strconv.ParseInt(tsEndStr, 10, 64)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			tsEnd = end
		}

		params := make(map[string]int)

		params["pageNum"] = int(pageNum)
		params["pageSize"] = int(pageSize)
		if uin != 0 {
			params["uin"] = int(uin)
		}
		if msgType != 0 {
			params["msgType"] = int(msgType)
		}
		if tsStart != 0 {
			params["tsStart"] = int(tsStart)
		}
		if tsEnd != 0 {
			params["tsEnd"] = int(tsEnd)
		}

		req := &GetAllStoryReq{params}

		rsp, err := doGetAllStorys(req)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		d, err := json.Marshal(rsp)
		if err != nil {
			log.Errorf("json marshal error", err.Error())
			return
		}

		io.WriteString(w, string(d))
	}

	log.Debugf("end GetAllStoryHandler")
	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("start Handler r:%+v", r)

	filePath := "../doc/" + r.URL.Path[1:]
	log.Debugf("Path:%s", filePath)
	http.ServeFile(w, r, filePath)
	log.Debugf("end Handler")
	return
}
