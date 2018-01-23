package addr

import (
	"common/constant"
	//"common/env"
	"common/mydb"
	"common/rest"
	"common/token"
	"common/util"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/cache"
	"svr/st"
	"time"
)

type UpdateAddrReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Data string `schema:"data"`
}

type UpdateAddrRsp struct {
	Cnt   int              `json:"cnt"`
	Infos []*Phone2UinInfo `json:"infos"`
}

type Phone2UinInfo struct {
	Phone      string `json:"phone"`
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	OrgPhone   string `json:"orgPhone"`
}

type AddrInfo struct {
	Phone string `schema:"phone"`
	Name  string `schema:"name"`
}

func (this *AddrInfo) String() string {
	return fmt.Sprintf(`AddrInfo{Phone:%s, Name:%s}`, this.Phone, this.Name)
}

func (this *Phone2UinInfo) String() string {
	return fmt.Sprintf(`Phone2UinInfo{Phone:%s, Uin:%d, OrgPhone:%s}`, this.Phone, this.Uin, this.OrgPhone)
}

func doUpdateAddr(req *UpdateAddrReq, r *http.Request) (rsp *UpdateAddrRsp, err error) {

	log.Errorf("uin %d, UpdateAddrReq %+v", req.Uin, req)

	uuid, err := token.GetUuidFromTokenString(req.Token, req.Ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt "+err.Error())
		log.Errorf("uin %d, UpdateAddrRsp error, %s", req.Uin, err.Error())
		return
	}

	cnt, infos, err := UpdateAddr(req.Uin, uuid, req.Data)
	if err != nil {
		log.Errorf("uin %d, UpdateAddrRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &UpdateAddrRsp{cnt, infos}

	log.Errorf("uin %d, UpdateAddrRsp succ, %+v", req.Uin, rsp)

	return
}

func UpdateAddr(uin int64, uuid int64, data string) (cnt int, infos []*Phone2UinInfo, err error) {

	infos = make([]*Phone2UinInfo, 0)

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

	if len(addrs) == 0 {
		log.Error("update addrs size 0")
		return
	}

	log.Errorf("uin %d, uuid %d, upload addr %+v", uin, uuid, addrs)

	/*
		if len(addrs) > env.Config.Addr.UploadBatchSize {
			err = rest.NewAPIError(constant.E_INVALID_PARAM, "addr batch size over limit")
			log.Error(err.Error())
			return
		}
	*/

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	stmt, err := inst.Prepare(`insert into addrBook(uuid, friendPhone, friendName, friendUin, uploaderUin, status, ts) values(?, ?, ?, ?, ?, ?, ?) on duplicate key update friendName = ?, friendUin = ?, uploaderUin = ?, status = ?, ts = ?`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	stmt1, err := inst.Prepare(`insert into addrBookOrg(id, uuid, orgPhone, name, phone, uploaderUin, ts) values(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt1.Close()

	friendUins := make([]int64, 0)

	ts := time.Now().Unix()
	for _, addr := range addrs {

		phone, valid := util.PhoneValid2(addr.Phone)

		if !valid {
			phone = ""
		}

		//插入到备份数据库 仅做保留分析
		_, err1 := stmt1.Exec(0, uuid, addr.Phone, addr.Name, phone, uin, ts)
		if err1 != nil {
			log.Errorf(err1.Error())
		}

		if !valid {
			continue
		}

		if len(addr.Name) == 0 {
			continue
		}

		friendUin := cache.PHONE2UIN[phone]

		if friendUin != 0 {
			friendUins = append(friendUins, friendUin)
		}

		status := 0
		_, err = stmt.Exec(uuid, phone, addr.Name, friendUin, uin, status, ts, addr.Name, friendUin, uin, status, ts)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}

		cnt += 1
		//先临时保存 后面还有校验是否注册完成
		infos = append(infos, &Phone2UinInfo{phone, friendUin, "", "", addr.Phone})
	}

	res, err := st.BatchGetUserProfileInfo(friendUins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for i, info := range infos {

		//只有注册完成的才返回
		if v, ok := res[info.Uin]; ok {

			if len(v.NickName) > 0 {
				//注册完成
				info.NickName = v.NickName
				info.HeadImgUrl = v.HeadImgUrl
			} else {
				//未注册完成
				info.Uin = 0
				info.NickName = ""
				info.HeadImgUrl = ""
			}

		} else {
			//未查到资料
			info.Uin = 0
			info.NickName = ""
			info.HeadImgUrl = ""
		}

		infos[i] = info
	}

	return
}
