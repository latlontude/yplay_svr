package rest

import (
	"fmt"
	"github.com/gorilla/context"
	"net/http"
)

type t_REDIRECT int

type RedirectInfo struct {
	RedirectUrl string `json:"redirectUrl"`
}

func (f t_REDIRECT) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	obj := context.Get(req, CTX_API_RESULT)

	rsp, ok := obj.(*RedirectInfo)
	if ok {

		http.Redirect(w, req, rsp.RedirectUrl, http.StatusSeeOther)

	} else {
		fmt.Println("json servehttp ", obj)
	}
}

const (
	REDIRECT = t_REDIRECT(0)
)
