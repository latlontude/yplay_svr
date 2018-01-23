package rest

import (
	"errors"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/schema"
	"net/http"
	"reflect"
	"unicode"
	"unicode/utf8"
)

const (
	CTX_API_RESULT = "rest.api.result"
)

type errorHandler struct {
	Err error
}

func (e errorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	panic(e.Err)
}

func errHandler(argv ...interface{}) http.Handler {
	return errorHandler{errors.New(fmt.Sprint(argv...))}
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

type ServeHttpInfo struct {
	W http.ResponseWriter
	R *http.Request
}

var (
	typeOfError            = reflect.TypeOf((*error)(nil)).Elem()
	typeOfHttpRequestPtr   = reflect.TypeOf((*http.Request)(nil))
	typeOfServeHttpInfoPtr = reflect.TypeOf((*ServeHttpInfo)(nil))
)

type restRpc struct {
	funv    reflect.Value
	argType reflect.Type
	reqType reflect.Type
	//replyType reflect.Type
}

// 将函数封装为一个http.Handler.
//
// fun必须为以下格式之一：
//   1. func() (*ReturnType, error)
//   2. func(req *http.Request) (*ReturnType, error)
//   3. func(params *ParamsType) (*ReturnType, error)
//   4. func(params *ParamsType, req *http.Request) (*ReturnType, error)
//
// ParamsType的字段必须按以下方式定义，才可以自动从QueryString接收参数：
//
//   type FooParams struct {
//      Foo int    `schema:"foo"`
//      Bar string `schema:"bar"`
//   }
func RPC(fun interface{}) http.Handler {
	funv := reflect.ValueOf(fun)
	funType := funv.Type()
	if funv.Kind() != reflect.Func {
		return errHandler("not a function: ", funType)
	}
	numIn := funType.NumIn()
	if numIn > 2 || funType.NumOut() != 2 {
		return errHandler("wrong signature: ", funType)
	}

	rpc := &restRpc{funv: funv}

	// last arg may be a *http.Request
	if numIn > 0 {
		lastArg := funType.In(numIn - 1)
		if lastArg == typeOfHttpRequestPtr || lastArg == typeOfServeHttpInfoPtr {
			rpc.reqType = lastArg
			numIn--
		}
	}
	if numIn == 1 {
		rpc.argType = funType.In(0)
		if !isExportedOrBuiltinType(rpc.argType) {
			return errHandler("argument type not exported: ", rpc.argType)
		}
	}

	if returnType := funType.Out(1); returnType != typeOfError {
		return errHandler("method", funType.Name(), "returns", returnType.String(), "not error")
	}
	return rpc
}

func (j *restRpc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var args []reflect.Value
	if j.argType != nil && j.argType.Kind() != reflect.Invalid {
		var argv reflect.Value
		argIsValue := false
		if j.argType.Kind() == reflect.Ptr {
			argv = reflect.New(j.argType.Elem())
		} else {
			argv = reflect.New(j.argType)
			argIsValue = true
		}

		err := decodeQueryParams(req, argv.Interface())
		if err != nil {
			if _, ok := err.(schema.MultiError); !ok { // 忽略MultiError:即忽略未定义的参数名
				panic(err)
			}
		}
		if argIsValue {
			argv = argv.Elem()
		}
		args = append(args, argv)
	}
	if j.reqType == typeOfHttpRequestPtr {
		args = append(args, reflect.ValueOf(req))
	} else if j.reqType == typeOfServeHttpInfoPtr {
		info := &ServeHttpInfo{w, req}
		args = append(args, reflect.ValueOf(info))
	}
	result := j.funv.Call(args)
	e := result[1].Interface()
	if e != nil {
		panic(e)
	}

	context.Set(req, CTX_API_RESULT, result[0].Interface())
}
