package rest

import (
	"fmt"
	"net/http"
)

// 带错误码的API错误对象 -------------------------------------------------------
type APIError struct {
	Code    int    //错误码
	Message string //错误信息
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[code:%d](%s)", e.Code, e.Message)
}

// TODO: 按照输出格式输出结果，目前只能输出json格式
func (e *APIError) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ret := &APIReply{
		Code:    e.Code,
		Msg:     e.Message,
		Payload: nil,
	}

	WriteJson(w, req, ret)
}

func NewAPIError(code int, msg string) *APIError {
	return &APIError{
		Code:    code,
		Message: msg,
	}
}
