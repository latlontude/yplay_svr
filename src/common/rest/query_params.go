package rest

import (
	"github.com/gorilla/schema"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var decoder = schema.NewDecoder()

func init() {
	decoder.RegisterConverter(time.Now(), convertTime)
}

func convertTime(s string) reflect.Value {
	t, err := time.Parse("2006-01-02 03:04:05", s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
	}
	if err != nil {
		return reflect.Value{}
	}
	return reflect.ValueOf(t)
}

// 获取一个请求参数. 根据输入类型进行转换。暂时支持string和int.
//    var page int
//    GetQueryParam(req, "page", &page)
func GetQueryParam(req *http.Request, name string, val interface{}) bool {
	s := req.FormValue(name)
	if s == "" {
		return false
	}
	switch v := val.(type) {
	case *string:
		*v = s
		return true
	case *int:
		if i, err := strconv.Atoi(s); err == nil {
			*v = i
			return true
		}
	}
	return false
}

func getQueryParams(req *http.Request, args ...interface{}) {
	cnt := len(args)
	if cnt%2 != 0 {
		cnt--
	}
	for i := 0; i < cnt/2; i++ {
		name, val := args[2*i], args[2*i+1]
		if s, ok := name.(string); ok {
			GetQueryParam(req, s, val)
		}
	}
}

func decodeQueryParams(req *http.Request, arg interface{}) error {
	req.FormValue("") // 触发ParseForm
	return decoder.Decode(arg, req.Form)
}
