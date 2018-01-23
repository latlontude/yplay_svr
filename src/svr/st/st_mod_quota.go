package st

import (
	"common/constant"
	"common/env"
	"common/mydb"
	"common/rest"
	"fmt"
	"time"
)

type ProfileModQuotaInfo struct {
	Uin       int64 `json:"uin"`
	Field     int   `json:"field"`
	HasModCnt int   `json:"hasModCnt"`
	LeftCnt   int   `json:"leftCnt"`
}

func (this *ProfileModQuotaInfo) String() string {

	return fmt.Sprintf(`ProfileModQuotaInfo{Uin:%d, Field:%d, HasModCnt:%d, LeftCnt:%d}`,
		this.Uin, this.Field, this.HasModCnt, this.LeftCnt)
}

func GetUserProfileModQuotaInfo(uin int64, field int) (info *ProfileModQuotaInfo, err error) {

	if uin == 0 || (field <= constant.ENUM_PROFILE_MOD_FIELD_MIN || field >= constant.ENUM_PROFILE_MOD_FIELD_MAX) {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param field")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	sql := fmt.Sprintf(`select uin, modField, count(modTs) from profileModRecords where uin = %d and modField = %d`, uin, field)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	info = &ProfileModQuotaInfo{uin, field, 0, 100000}

	if info.Field == constant.ENUM_PROFILE_MOD_FIELD_NICKNAME || info.Field == constant.ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE {
		info.LeftCnt = env.Config.Profile.ModMaxCnt
	}

	if info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER || info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {
		info.LeftCnt = env.Config.Profile.GenderModMaxCnt
	}

	for rows.Next() {

		rows.Scan(
			&info.Uin,
			&info.Field,
			&info.HasModCnt,
		)

		//只有修改nickname和学校才会限制修改次数
		if info.Field == constant.ENUM_PROFILE_MOD_FIELD_NICKNAME ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {

			info.LeftCnt = env.Config.Profile.ModMaxCnt - info.HasModCnt

			if info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER || info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {
				info.LeftCnt = env.Config.Profile.GenderModMaxCnt - info.HasModCnt
			}

		} else {

			info.LeftCnt = 100000

		}

		if info.LeftCnt < 0 {
			info.LeftCnt = 0
		}

		break
	}

	return
}

func GetUserProfileModQuotaAllInfo(uin int64) (infos []*ProfileModQuotaInfo, err error) {

	infos = make([]*ProfileModQuotaInfo, 0)

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}
	sql := fmt.Sprintf(`select uin, modField, count(modTs) from profileModRecords where uin = %d group by modField`, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var info ProfileModQuotaInfo

		rows.Scan(
			&info.Uin,
			&info.Field,
			&info.HasModCnt,
		)

		//只有修改nickname和学校/userName/gender才会限制修改次数
		if info.Field == constant.ENUM_PROFILE_MOD_FIELD_NICKNAME ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {

			info.LeftCnt = env.Config.Profile.ModMaxCnt - info.HasModCnt

			if info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER || info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {
				info.LeftCnt = env.Config.Profile.GenderModMaxCnt - info.HasModCnt
			}

		} else {
			info.LeftCnt = 100000
		}

		if info.LeftCnt < 0 {
			info.LeftCnt = 0
		}

		infos = append(infos, &info)
	}

	for field := constant.ENUM_PROFILE_MOD_FIELD_MIN + 1; field < constant.ENUM_PROFILE_MOD_FIELD_MAX; field++ {

		find := false
		for _, v := range infos {

			if v.Field == field {
				find = true
				break
			}
		}

		if find {
			continue
		}

		info := &ProfileModQuotaInfo{uin, field, 0, 100000}

		if field == constant.ENUM_PROFILE_MOD_FIELD_NICKNAME ||
			field == constant.ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER ||
			info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {

			info.LeftCnt = env.Config.Profile.ModMaxCnt

			if info.Field == constant.ENUM_PROFILE_MOD_FIELD_GENDER || info.Field == constant.ENUM_PROFILE_MOD_FIELD_USERNAME {
				info.LeftCnt = env.Config.Profile.GenderModMaxCnt
			}
		}

		infos = append(infos, info)
	}

	return
}

func AddProfileModRecordInfo(uin int64, field int, desc string) (err error) {

	if uin == 0 || (field <= constant.ENUM_PROFILE_MOD_FIELD_MIN || field >= constant.ENUM_PROFILE_MOD_FIELD_MAX) {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param field")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	stmt, err := inst.Prepare(`insert ignore into profileModRecords values(?,?,?,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	_, err = stmt.Exec(uin, field, ts, desc)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	return
}

func AddMultiProfileModRecordInfo(uin int64, mods map[int]string) (err error) {

	if uin == 0 || len(mods) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param field")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		return
	}

	stmt, err := inst.Prepare(`insert ignore into profileModRecords values(?,?,?,?)`)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_PREPARE, err.Error())
		log.Error(err.Error())
		return
	}
	defer stmt.Close()

	ts := time.Now().Unix()

	for field, desc := range mods {

		if field <= constant.ENUM_PROFILE_MOD_FIELD_MIN || field >= constant.ENUM_PROFILE_MOD_FIELD_MAX {
			continue
		}

		_, err = stmt.Exec(uin, field, ts, desc)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}
	}

	return
}
