package express

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func GetSign(orderInfo interface{}, secret string) (sign string) {
	paramsMap := make(map[string]interface{}, 0)
	object := reflect.ValueOf(orderInfo)
	ref := object.Elem()
	typeOfType := ref.Type()
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Field(i)
		tag := typeOfType.Field(i).Tag.Get("xml")
		if field.Interface() != "" && tag != "xml" {

			//过滤掉int型参数为0的字段
			if field.Type().Name() == "int" && field.Int() == 0 {
				continue
			}
			//去掉sign
			if field.Type().Name() == "sign" {
				continue
			}
			paramsMap[tag] = field.Interface()
		}
	}

	//按照key字典顺序排序
	keys := make([]string, 0)
	for k := range paramsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	var stringA string
	for _, key := range keys {
		stringA += fmt.Sprintf("%s=%s&", key, paramsMap[key])
	}
	//最后拼接秘钥
	stringA += fmt.Sprintf("%s=%s", "key", "4sbBcRjiGtWMSsOt56JYR1knwkTWe5TU")
	log.Debugf("%s", stringA)

	//md5
	md5Ctx2 := md5.New()
	md5Ctx2.Write([]byte(stringA))
	sign = hex.EncodeToString(md5Ctx2.Sum(nil))
	sign = strings.ToUpper(sign) //签名

	return
}
