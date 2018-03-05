package rest

import (
	//"fmt"
	"github.com/gorilla/context"
	"io"
	"net/http"
)

// RAW Handler.
// 将context["API.RESULT"]中的数据以二进制格式输出
type t_RAW int

func (f t_RAW) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	obj := context.Get(req, CTX_API_RESULT)

	switch obj.(type) {

	default:
		stringData := obj.(*string)

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/octstream;charset=utf-8")
		w.Write([]byte(*stringData))

		io.WriteString(w, "\n")
	}
}

const (
	RAW = t_RAW(0) // JSON输出Handler
)
