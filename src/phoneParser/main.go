package main

import (
	"flag"
	"fmt"
	"github.com/ttacon/libphonenumber"
)

var (
	phoneStr string
	country  string
)

func init() {
	flag.StringVar(&phoneStr, "s", "", "电话号码")
	flag.StringVar(&country, "c", "", "国家代码")
}

func main() {

	flag.Parse()

	phone, err := libphonenumber.Parse(phoneStr, country)
	if err != nil {
		fmt.Printf("Parse [%s] error %s\n", phoneStr, err.Error())
		return
	}

	fmt.Printf("%+v\n", phone)

	fmt.Printf("phoneStr:%d", phone.GetNationalNumber())

	return
}
