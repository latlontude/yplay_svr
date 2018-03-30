package rest

import (
	"fmt"
	"github.com/gorilla/context"
	"net/http"
	"os"
)

type t_DOWNLOAD int

type DownloadInfo struct {
	Uin          int64  `json:"uin"`
	Token        string `json:"token"`
	Ver          int    `json:"ver"`
	DownloadFile string `json:"downloadFile"`
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (f t_DOWNLOAD) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	obj := context.Get(req, CTX_API_RESULT)

	di, ok := obj.(*DownloadInfo)
	if ok {

		path := fmt.Sprintf("../download/%s", di.DownloadFile)
		exist, _ := PathExists(path)
		if !exist {
			fmt.Println("file not exist ", path)
			return
		}

		if di.Uin > 0 {

			ck1 := http.Cookie{Name: "uin", Value: fmt.Sprintf("%d", di.Uin), Path: "/"}
			ck2 := http.Cookie{Name: "token", Value: di.Token, Path: "/"}
			ck3 := http.Cookie{Name: "ver", Value: fmt.Sprintf("%d", di.Ver), Path: "/"}

			http.SetCookie(w, &ck1)
			http.SetCookie(w, &ck2)
			http.SetCookie(w, &ck3)

			http.ServeFile(w, req, path)
		} else {
			fmt.Println("uin not found ", di.Uin)
		}

	} else {
		fmt.Println("json servehttp ", obj)
	}
}

const (
	DOWNLOAD = t_DOWNLOAD(0)
)
