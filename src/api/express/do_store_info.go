//存储下单地址
package express

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"net/http"
)

type StoreInfoReq struct {
	Openid   string `schema:"openid"`
	SchoolId int    `schema:"schoolId"`
	Name     string `schema:"name"`
	Phone    string `schema:"phone"`
	Address  string `schema:"address"`
}

type StoreInfoRsp struct {
	Code int `json:"code"`
}

func doStoreInfo(req *StoreInfoReq, r *http.Request) (rsp *StoreInfoRsp, err error) {
	log.Debugf("GetPositionListReq:%+v", req)

	err = StoreInfo(req.Openid, req.SchoolId, req.Name, req.Phone, req.Address)
	if err != nil {
		log.Errorf("GetPositionList error,err:%+v", err)
	}

	code := 0
	rsp = &StoreInfoRsp{code}
	log.Debugf("StoreInfoRsp : %+v", rsp)
	return
}

//获取某个学校的位置信息
func StoreInfo(openid string, schoolId int, name string, phone string, address string) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	stmt, err := inst.Prepare(`replace into express_userinfo(openid, schoolId, name, phone,address) 
		values(?, ?, ?, ?, ?)`)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(openid, schoolId, name, phone, address)

	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	_, err = res.LastInsertId()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}
	return
}
