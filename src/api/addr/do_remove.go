package addr

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"common/token"
	"common/util"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type RemoveAddrReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Data string `schema:"data"`
}

type RemoveAddrRsp struct {
	Cnt int `json:"cnt"`
}

func doRemoveAddr(req *RemoveAddrReq, r *http.Request) (rsp *RemoveAddrRsp, err error) {

	log.Errorf("uin %d, RemoveAddrReq %+v", req.Uin, req)

	t, err := token.DecryptToken(req.Token, req.Ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt fail")
		log.Errorf("uin %d, RemoveAddrRsp error, %s", req.Uin, err.Error())
		return
	}

	cnt, err := RemoveAddr(req.Uin, t.Uuid, req.Data)
	if err != nil {
		log.Errorf("uin %d, RemoveAddrRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &RemoveAddrRsp{cnt}

	log.Errorf("uin %d, RemoveAddrRsp succ, %+v", req.Uin, rsp)

	return
}

func RemoveAddr(uin int64, uuid int64, data string) (cnt int, err error) {

	if uin == 0 || len(data) == 0 || uuid < constant.ENUM_DEVICE_UUID_MIN {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	cnt = 0
	decodeData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err.Error())
		return
	}

	var addrs []AddrInfo
	err = json.Unmarshal([]byte(decodeData), &addrs)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err.Error())
		return
	}

	log.Errorf("uin %d, uuid %d, RemoveAddr %+v", uin, uuid, addrs)

	if len(addrs) == 0 {
		log.Error("update addrs size 0")
		return
	}

	if len(addrs) > constant.ENUM_ADDR_MAX_BATCH_UPLOAD_SIZE {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "addr batch size over limit")
		log.Error(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	stmt, err := inst.Prepare(`delete from addrBook where uuid = ? and friendPhone = ?`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	for _, addr := range addrs {

		phone, valid := util.PhoneValid(addr.Phone)
		if !valid {
			continue
		}

		cnt += 1

		_, err = stmt.Exec(uuid, phone)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}
	}

	return
}
