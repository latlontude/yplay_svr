package account

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"common/sms"
	"fmt"
	"net/http"
	"time"
)

type CheckInviteCodeReq struct {
	Phone      string `schema:"phone"`
	InviteCode string `schema:"inviteCode"`
}

type CheckInviteCodeRsp struct {
}

func doCheckInviteCode(req *CheckInviteCodeReq, r *http.Request) (rsp *CheckInviteCodeRsp, err error) {

	log.Debugf("phone %s, CheckInviteCodeReq %+v", req.Phone, req)

	err = CheckInviteCode(req.Phone, req.InviteCode)
	if err != nil {
		log.Errorf("phone %s, CheckInviteCodeRsp error %s", req.Phone, err.Error())
		return
	}

	rsp = &CheckInviteCodeRsp{}

	log.Debugf("phone %s, CheckInviteCodeRsp succ, %+v", req.Phone, rsp)

	return
}

func CheckInviteCode(phone string, inviteCode string) (err error) {

	if len(phone) == 0 || len(inviteCode) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Errorf(err.Error())
		return
	}

	if !sms.IsValidPhone(phone) {
		err = rest.NewAPIError(constant.E_INVALID_PHONE, "phone number invalid")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_INVITE_CODE)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%s", inviteCode)
	exist, err := app.Exist(keyStr) //inviteCode 缓存7天
	if err != nil {
		err = rest.NewAPIError(constant.E_SMS_CODE_ERR, "code not exist in redis")
		log.Errorf(err.Error())
		return
	}

	if !exist {
		err = rest.NewAPIError(constant.E_INVITE_CODE_INVALID, "invalid invite code")
		log.Errorf(err.Error())
		return
	}

	//添加到数据库中
	err = AddPhoneToInviteList(phone, inviteCode)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//从内存中删掉
	app.Del(keyStr)

	return
}

func AddPhoneToInviteList(phone string, inviteCode string) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	q := fmt.Sprintf(`insert ignore into invitecode values("%s", "%s", %d)`, phone, inviteCode, time.Now().Unix())

	_, err = inst.Exec(q)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return

}
