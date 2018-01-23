package util

import (
	"fmt"
	"github.com/ttacon/libphonenumber"
	"strings"
)

func PhoneValid(org string) (phone string, isValid bool) {

	isValid = false

	//(086)13500000000 (+86)13500000000 (86)13500000000
	phone = strings.TrimSpace(org)
	phone = strings.Replace(phone, " ", "", -1) //去除空格

	//美国手机号
	if strings.HasPrefix(phone, "+1(") {
		return
	}

	phone = strings.Replace(phone, "(", "", -1) //去除括号
	phone = strings.Replace(phone, ")", "", -1) //去除括号

	//+86 13500000000
	if strings.HasPrefix(phone, "+86") {
		phone = phone[3:]
	}

	//086 13500000000
	if strings.HasPrefix(phone, "086") {
		phone = phone[3:]
	}

	//86 13500000000
	if strings.HasPrefix(phone, "86") {
		phone = phone[2:]
	}

	//last check phone num
	if len(phone) != 11 {
		return
	}

	//手机号码都是以1开头
	if string(phone[0]) != "1" {
		return
	}

	isValid = true

	return
}

func PhoneValid2(org string) (phone string, isValid bool) {

	isValid = false

	phoneSt, err := libphonenumber.Parse(org, "CN")
	if err != nil {
		return
	}

	phoneNum := phoneSt.GetNationalNumber()

	phone = fmt.Sprintf("%d", phoneNum)

	if len(phone) != 11 {
		return
	}

	if string(phone[0]) != "1" {
		return
	}

	isValid = true

	return
}
