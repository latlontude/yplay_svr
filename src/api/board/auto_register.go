package board

import (
	"api/account"
	"api/geneqids"
	"api/sns"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func GetPhoneMax() (phone int, err error) {
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select phone from profiles where locate('1406666',phone) order by ts desc limit 1`)
	rows, err1 := inst.Query(sql)
	if err1 != nil {
		err = err1
	}
	for rows.Next() {
		rows.Scan(&phone)
	}
	return
}

//每个学校注册小号

func AutoRegister(uInfo *st.UserProfileInfo) (registerUin []int64, err error) {
	registerUin = make([]int64, 0)
	phone, err := GetPhoneMax()
	phone = phone + 1
	if err != nil {
		log.Debugf("Get phone max error")
	}
	code := "0000"
	var now int64 = 1534214203000

	device := "PIC-AL00"
	os := "android_8.0.0_26"
	appVer := "1.0.4499_2018-08-09 19:32:23_beta"

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}



	sql := fmt.Sprintf(`select uin from profiles where locate('14066660',phone) and  schoolId = %d`,uInfo.SchoolId)
	rows, err1 := inst.Query(sql)
	if err1 != nil {
		err = err1
	}

	tmpUinList := make([]int64, 0)
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		tmpUinList = append(tmpUinList,uid)
	}

	//已经存在五个小号 就不要注册了
	if len(tmpUinList) >= 5 {
		registerUin = tmpUinList
		return
	}

	var nickName, headImgUrl, country, province, schoolName, deptName string
	var gender, age, grade, enrollmentYear, deptId int
	sql = fmt.Sprintf(`select nickName,headImgUrl,gender,age,grade,deptId,deptName,country,enrollmentYear 
from profiles where locate('14066660',phone) and length(headImgUrl)>0 ORDER BY RAND() limit 5`)
	rows, err1 = inst.Query(sql)
	if err1 != nil {
		err = err1
	}

	schoolName = uInfo.SchoolName
	province = uInfo.Province

	for rows.Next() {
		rows.Scan(&nickName, &headImgUrl, &gender, &age, &grade, &deptId, &deptName, &country, &enrollmentYear)
		phoneStr := fmt.Sprintf("%d", phone)
		rsp, rspErr := account.Login2(phoneStr, code, now, device, os, appVer)
		if rspErr != nil {
			err = rspErr
			return
		}

		registerUin = append(registerUin, rsp.Uin)

		updateSql := fmt.Sprintf(`update profiles set 
					nickName  ='%s',
					headImgUrl = '%s',
					gender = %d,
					age = %d,
					grade = %d,
					schoolId = %d,
					schoolType = 3 ,
					schoolName='%s',
					deptId = %d , 
					deptName = '%s',
					country = '%s',
					province = '%s',
					enrollmentYear = %d 
					where uin = %d
					`, nickName, headImgUrl, gender, age, grade, uInfo.SchoolId, schoolName, deptId, deptName, country, province, enrollmentYear, rsp.Uin)

		_, err = inst.Exec(updateSql)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Errorf(err.Error())
			return
		}
		log.Debugf("sql:%s", updateSql)

		//表示资料注册完成了，这时再更新通讯录里面的状态
		if len(nickName) > 0 {
			//添加客服号
			var serviceAccountUin int64
			serviceAccountUin = 100001 //客服号
			sns.AddCustomServiceAccount(rsp.Uin, serviceAccountUin)
			//14066660301 客服号  102688
			sns.AddCustomServiceAccount(rsp.Uin, 102688)
		}

		//修改性别重新生成答题列表
		if gender >= 1 && gender <= 2 {
			//重新生成答题列表
			go geneqids.Gene(rsp.Uin)
		}
		phone++
	}
	return

}
