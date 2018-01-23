package rest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 向URL发起HTTP GET请求，返回的JSON结果转换为相应对象.
func GetJSON(baseUrl string, params url.Values, ret interface{}) error {
	resp, err := http.Get(baseUrl + "?" + params.Encode())
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, ret)
}

func WriteJson(w http.ResponseWriter, req *http.Request, obj interface{}) error {
	pretty := req.FormValue("_pretty_")

	var data []byte
	var err error
	if pretty != "" {
		data, err = json.MarshalIndent(obj, "", "  ")
	} else {
		data, err = json.Marshal(obj)
	}
	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(data)

	io.WriteString(w, "\n")
	return nil
}
