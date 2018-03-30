package httputil

import (
	"common/rest"
	"github.com/gorilla/mux"
	"net/http"
	"path"
)

var (
	Router         *mux.Router
	HttpPathPrefix string
)

func Init() (err error) {

	router := mux.NewRouter()

	HttpPathPrefix = "/"

	Router = router.PathPrefix(HttpPathPrefix).Subrouter()

	return
}

func HandleStatic(urlPrefix, relPath string) {

	fs0 := http.FileServer(http.Dir(relPath))

	n := len(urlPrefix)
	if n == 0 || urlPrefix[n-1] != '/' {
		urlPrefix += "/"
	}

	fs := http.StripPrefix(path.Join(HttpPathPrefix, urlPrefix), fs0)
	fs = rest.HandlerChain{
		fs,
	}

	Router.PathPrefix(urlPrefix).Handler(fs)
}

type APIMap map[string]http.Handler

func HandleAPIMap(urlPrefix string, apiMap APIMap) {

	r := Router.Get(urlPrefix)
	if r == nil {
		r = Router.PathPrefix(urlPrefix).Name(urlPrefix)
	}

	sub := r.Subrouter()
	for pattern, fun := range apiMap {
		sub.Handle(pattern, fun)
	}
}

//bindAddr -> 10.154.216.215:9091
func ListenHttp(bindAddr string) error {

	var handler http.Handler

	handler = Router

	return http.ListenAndServe(bindAddr, handler)
}
