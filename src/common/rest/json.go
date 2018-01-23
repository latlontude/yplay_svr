package rest

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"github.com/gorilla/context"
	"io/ioutil"
	"net/http"
)

// JSON Handler.
// 将context["API.RESULT"]中的数据以JSON(或JSONP)格式输出
type t_JSON int

func (f t_JSON) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	obj := context.Get(req, CTX_API_RESULT)

	//fmt.Println("json servehttp ", obj)

	switch c := obj.(type) {
	case chan interface{}: // TODO: 尚不支持server push
		for v := range c {
			err := WriteJson(w, req, v)
			if err != nil {
				panic(err)
			}
		}
	default:
		obj1 := &APIReply{Code: 0, Msg: "succ", Payload: c}

		//fmt.Println("write json ", obj1)

		err := WriteJson(w, req, obj1)

		if err != nil {
			panic(err)
		}
	}
}

const (
	JSON = t_JSON(0) // JSON输出Handler
)

func RequestJson(url string, req, ret interface{}) (err error) {

	json_string, err := json.Marshal(req)

	resp, err := http.Post(url, "application/json", bytes.NewReader(json_string))
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, ret)
	return
}
