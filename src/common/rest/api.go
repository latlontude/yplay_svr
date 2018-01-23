// HTTP RESTful API库.
//
// 提供URL
package rest

import (
	"fmt"
	"net/http"
)

type HandlerChain []http.Handler

func (chain HandlerChain) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {

			switch v := r.(type) {
			case http.Handler: // 如果err对象实现了http.Handler接口，则调用其输出函数。
				v.ServeHTTP(w, req)
			case error: // 普通err对象，打印出stack trace.
				// buf := make([]byte, 4096)
				// n := runtime.Stack(buf, false)
				// msg := fmt.Sprintf("Internal Server Error: %s\n\n%s\n", v.Error(), string(buf[:n]))
				// log.Error(req.RequestURI, " ", v.Error())
				// log.Error(string(buf[:n]))

				msg := fmt.Sprintf("Internal Server Error: %s", v.Error())
				// hij, ok := w.(http.Hijacker)
				// if !ok {
				// 	fmt.Println(v)
				http.Error(w, msg, http.StatusInternalServerError)
				// } else {
				// 	fmt.Println(hij)
				// }

			}
		}
	}()

	for _, filter := range chain {
		filter.ServeHTTP(w, req)
	}
}

// 通用返回结果格式。
type APIReply struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Payload interface{} `json:"payload,omitempty"`
}

/*
func Apify2(fun interface{}) http.Handler {
	return HandlerChain{
		RPC(fun),
		JSON,
	}
}

func Apify3(fun interface{}) http.Handler {
	return HandlerChain{
		//AUTH,
		RPC(fun),
		JSON,
	}
}
*/
