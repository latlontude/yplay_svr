package addr

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"common/util"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"svr/st"
)

type QueryByPhoneReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Data string `schema:"data"`
}

type QueryByPhoneRsp struct {
	Infos []*st.UserProfileInfo `json:"infos"`
}

func doQueryByPhone(req *RemoveAddrReq, r *http.Request) (rsp *QueryByPhoneRsp, err error) {

	log.Errorf("uin %d, QueryByPhoneReq %+v", req.Uin, req)

	infos, err := QueryByPhone(req.Uin, req.Data)
	if err != nil {
		log.Errorf("uin %d, QueryByPhoneRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &QueryByPhoneRsp{infos}

	log.Errorf("uin %d, QueryByPhoneRsp succ, %+v", req.Uin, rsp)

	return
}

func QueryByPhone(uin int64, data string) (infos []*st.UserProfileInfo, err error) {

	infos = make([]*st.UserProfileInfo, 0)

	if uin == 0 || len(data) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	decodeData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err.Error())
		return
	}

	var phones []string
	err = json.Unmarshal([]byte(decodeData), &phones)
	if err != nil {
		err = rest.NewAPIError(constant.E_DECODE_ERR, err.Error())
		log.Error(err.Error())
		return
	}

	log.Debugf("uin %d, QueryByPhone %+v", uin, phones)

	if len(phones) == 0 {
		log.Errorf("query phones size 0")
		return
	}

	if len(phones) > constant.ENUM_ADDR_MAX_BATCH_UPLOAD_SIZE {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "phone batch size over limit")
		log.Errorf(err.Error())
		return
	}

	nphones := make([]string, 0)
	phoneStr := ""

	for _, phone := range phones {

		phone, valid := util.PhoneValid(phone)
		if !valid {
			continue
		}

		nphones = append(nphones, phone)

		phoneStr += "\"" + phone + "\"" + ","
	}

	if len(nphones) == 0 {
		log.Errorf("invalid query phones size 0")
		return
	}
	phoneStr = phoneStr[:len(phoneStr)-1]

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, userName, phone, nickName, headImgUrl, gender, age, grade, schoolId, schoolType, schoolName, deptId, deptName, country, province, city from profiles where status = 1 and phone in (%s)`, phoneStr)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		info := &st.UserProfileInfo{}

		rows.Scan(
			&info.Uin,
			&info.UserName,
			&info.Phone,
			&info.NickName,
			&info.HeadImgUrl,
			&info.Gender,
			&info.Age,
			&info.Grade,
			&info.SchoolId,
			&info.SchoolType,
			&info.SchoolName,
			&info.DeptId,
			&info.DeptName,
			&info.Country,
			&info.Province,
			&info.City)

		if len(info.HeadImgUrl) > 0 {
			info.HeadImgUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/headimgs/%s", info.HeadImgUrl)
		}

		if len(info.NickName) > 0 {
			infos = append(infos, info)
		}
	}

	return
}
