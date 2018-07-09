package user

import (
	"api/geneqids"
	"api/sns"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"strings"
	"svr/st"
)

type UpdateUserProfileReq struct {
	Uin       int64  `schema:"uin"`
	Token     string `schema:"token"`
	Ver       int    `schema:"ver"`
	NickName  string `schema:"nickName"`
	HeadImgId string `schema:"headImgId"`
	Gender    int    `schema:"gender"`
	UserName  string `schema:"userName"`
	Age       int    `schema:"age"`

	//没有省 市 家乡
	Country         string `json:"country"`
	Province        string `json:"province"`
	City            string `json:"city"`
	Hometown        string `json:"hometown"`        //家乡

	Flag      int    `schema:"flag"` //是否记入修改次数限制
}

type UpdateUserProfileRsp struct {
}

func doUpdateUserProfile(req *UpdateUserProfileReq, r *http.Request) (rsp *UpdateUserProfileRsp, err error) {

	log.Errorf("uin %d, UpdateUserProfileReq %+v", req.Uin, req)

	err = UpdateUserProfileInfo(req.Uin, req.NickName, req.HeadImgId, req.Gender, req.Age, req.UserName, req.Country,req.Province,req.City,req.Hometown,req.Flag)
	if err != nil {
		log.Errorf("uin %d, UpdateUserProfileRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &UpdateUserProfileRsp{}

	log.Errorf("uin %d, UpdateUserProfileRsp succ, %+v", req.Uin, rsp)

	return
}

//限制修改各个字段的修改次数
func UpdateUserProfileInfo(uin int64, nickName, headImgId string, gender, age int, userName ,country ,province ,city , hometown string, flag int) (err error) {

	log.Debugf("args %s,%s,%s,%s",country,province,city,hometown)
	if uin == 0 {
		return
	}

	if len(nickName) == 0 && len(headImgId) == 0 && (gender <= 0 || gender > 2) && len(userName) == 0 &&
		(age == 0) && len(country) == 0 && len(province) == 0 && len(city) == 0 &&len(hometown) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Error(err.Error())
		return
	}

	var modQutoaInfos []*st.ProfileModQuotaInfo

	if flag > 0 {
		modQutoaInfos, err = st.GetUserProfileModQuotaAllInfo(uin)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	//记录修改操作记录
	modsM := make(map[int]string)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	var sql string

	if len(userName) > 0 {

		sql = fmt.Sprintf(`select uin from profiles where userName = ?`)

		rows, err1 := inst.Query(sql, userName)
		if err1 != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err1.Error())
			log.Error(err.Error())
			return
		}
		defer rows.Close()

		find := false
		for rows.Next() {
			var id int
			rows.Scan(&id)
			find = true
		}

		if find {
			err = rest.NewAPIError(constant.E_USER_NAME_EXIST, "user name already exist!")
			log.Error(err.Error())
			return
		}
	}

	args := make([]interface{}, 0)

	sql = fmt.Sprintf(`update profiles set `)

	//nick太长
	if len(nickName) > 50 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "nick name too long")
		log.Debugf("nick len = %d",len(nickName))
		return
	}

	if len(nickName) > 0 && len(nickName) < 50 {

		if flag == 0 {
			sql += fmt.Sprintf(` nickName = ?, status = 1,`) //注册期间 设置昵称后表示资料完整了,注册流程中最后一步就是设置昵称
		} else {
			sql += fmt.Sprintf(` nickName = ?,`)
		}

		args = append(args, nickName)

		sensitiveStr := env.Config.Sensitive.Set // 获取配置文件中的敏感词库
		set := strings.Split(sensitiveStr, ",")
		log.Debugf("set :%+v", set)

		if flag > 0 {

			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {

				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_NICKNAME && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}

			for _, word := range set {
				if strings.Contains(nickName, word) {
					prompt := fmt.Sprintf("nickName contains sensitive word:%s", word)
					err = rest.NewAPIError(constant.E_INVALID_NICKNAME, prompt)
					log.Error(err.Error())
					return
				}
			}

			modsM[constant.ENUM_PROFILE_MOD_FIELD_NICKNAME] = fmt.Sprintf("nickName:%s", nickName)
		} else {
			for _, word := range set {
				if strings.Contains(nickName, word) {
					prompt := fmt.Sprintf("nickName contains sensitive word:%s", word)
					err = rest.NewAPIError(constant.E_INVALID_NICKNAME, prompt)
					log.Error(err.Error())
					return
				}
			}
		}
	}

	if len(userName) > 0 {
		sql += fmt.Sprintf(` userName = ?,`)
		args = append(args, userName)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {

				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}

			modsM[constant.ENUM_PROFILE_MOD_FIELD_USERNAME] = fmt.Sprintf("userName:%s", userName)
		}
	}

	if len(userName) > 0 {
		sql += fmt.Sprintf(` userName = ?,`)
		args = append(args, userName)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {

				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}

			modsM[constant.ENUM_PROFILE_MOD_FIELD_USERNAME] = fmt.Sprintf("userName:%s", userName)
		}
	}

	//只记录头像ID，下发的时候拼接域名
	if len(headImgId) > 0 {
		headImgUrl := fmt.Sprintf("%s", headImgId)
		sql += fmt.Sprintf(` headImgUrl = ?,`)
		args = append(args, headImgUrl)
	}

	if gender >= 1 && gender <= 2 {
		sql += fmt.Sprintf(` gender = ?,`)
		args = append(args, gender)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {

				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}

			modsM[constant.ENUM_PROFILE_MOD_FIELD_GENDER] = fmt.Sprintf("gender:%d", gender)
		}
	}

	if age > 0 && age < 100 {
		sql += fmt.Sprintf(` age = ?,`)
		args = append(args, age)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {

				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_AGE && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}

			modsM[constant.ENUM_PROFILE_MOD_FIELD_AGE] = fmt.Sprintf("age:%d", age)
		}
	}

	if len(country) > 0 {
		sql += fmt.Sprintf(` country = ?,`)
		args = append(args, country)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {
				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_COUNTRY && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}
			modsM[constant.ENUM_PROFILE_MOD_FIELD_COUNTRY] = fmt.Sprintf("country:%s", country)
		}
	}

	if len(province) > 0 {
		sql += fmt.Sprintf(` province = ?,`)
		args = append(args, province)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {
				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_PROVINCE && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}
			modsM[constant.ENUM_PROFILE_MOD_FIELD_PROVINCE] = fmt.Sprintf("province:%s", province)
		}
	}

	if len(city) > 0 {
		sql += fmt.Sprintf(` city = ?,`)
		args = append(args, city)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {
				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_CITY && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}
			modsM[constant.ENUM_PROFILE_MOD_FIELD_CITY] = fmt.Sprintf("city:%s", city)
		}
	}

	if len(hometown) > 0 {
		sql += fmt.Sprintf(` hometown = ?,`)
		args = append(args, hometown)

		if flag > 0 {
			//检查修改次数是否超限制
			for _, qi := range modQutoaInfos {
				if qi.Field == constant.ENUM_PROFILE_MOD_FIELD_HOMETOWN && qi.LeftCnt <= 0 {
					err = rest.NewAPIError(constant.E_PERMI_DENY, "modify cnt over limit!")
					log.Error(err.Error())
					return
				}
			}
			modsM[constant.ENUM_PROFILE_MOD_FIELD_HOMETOWN] = fmt.Sprintf("hometown:%s", hometown)
		}
	}


	sql = sql[:len(sql)-1]

	sql += fmt.Sprintf(` where uin = ?`)
	args = append(args, uin)


	//如果args为空 mysql报错
	if len(args) == 0 {
		return
	}


	log.Debugf("update file sql = %s",sql)
	_, err = inst.Exec(sql, args...)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//记入修改次数限制
	if flag > 0 {

		if len(modsM) > 0 {
			go st.AddMultiProfileModRecordInfo(uin, modsM)
		}
	}

	//表示资料注册完成了，这时再更新通讯录里面的状态
	if len(nickName) > 0 && flag == 0 {
		UpdateAddrBookInfo(uin)

		//预先生成答题列表
		geneqids.Gene(uin)

		//添加客服号
		sns.AddCustomServiceAccount(uin)
	}

	//修改性别重新生成答题列表
	if gender >= 1 && gender <= 2 && flag > 0 {
		//重新生成答题列表
		go geneqids.Gene(uin)
	}

	return
}

//注册完成之后再更新通讯录的相关好友UIN信息
func UpdateAddrBookInfo(uin int64) (err error) {

	if uin == 0 {
		return
	}

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(ui.Phone) < 5 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	q := fmt.Sprintf(`update addrBook set friendUin = %d where friendPhone = "%s"`, uin, ui.Phone)

	_, err = inst.Exec(q)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}
