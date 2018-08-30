package user

import (
	"api/account"
	"api/geneqids"
	"api/sns"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

/**
	给每个学校增加40个小号
 */
type AutoRegisterReq struct {
	Uin     int64  `schema:"uin"`
	Token   string `schema:"token"`
	Ver     int    `schema:"ver"`
	BoardId int    `schema:"boardId"`
}

type AutoRegisterRsp struct {
	UinArr []int64 `json:"uin_arr"`
}

//
func doAutoRegister(req *AutoRegisterReq, r *http.Request) (rsp *AutoRegisterRsp, err error) {
	//去除首位空白字符
	uidArr, err := AutoRegister()
	if err != nil {
		log.Debugf("AutoRegisterRsp err : %+v", err)
		return
	}
	rsp = &AutoRegisterRsp{uidArr}

	log.Debugf("AutoRegisterRsp : %+v", rsp)

	return

}
func AutoRegister() (registerUin []int64, err error) {

	registerUin = make([]int64, 0)
	var schools = []int{78811, 77557, 78976, 80561, 79238, 78964, 80560, 80562, 78627}
	//schoolMap[1000] = 14066661000

	phone := 14066665000
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

	var schoolMap = make(map[int]map[string]string, 0)
	//东莞理工学院     78811
	schoolMap[78811] = map[string]string{
		"schoolName": "东莞理工学院",
		"province":   "广东",
	}
	//黑龙江大学       77557
	schoolMap[77557] = map[string]string{
		"schoolName": "黑龙江大学",
		"province":   "黑龙江",
	}
	//湖南科技学院      78976
	schoolMap[ 78976] = map[string]string{
		"schoolName": "湖南科技学院 ",
		"province":   "湖南",
	}
	//东华理工大学       80561
	schoolMap[80561] = map[string]string{
		"schoolName": "东华理工大学",
		"province":   "江西",
	}
	//成都理工大学       79144
	schoolMap[79238] = map[string]string{
		"schoolName": "成都理工大学工程技术学院",
		"province":   "四川",
	}
	//湘潭大学          78964
	schoolMap[78964] = map[string]string{
		"schoolName": "湘潭大学",
		"province":   "湖南",
	}

	//云南财经          80560
	schoolMap[80560] = map[string]string{
		"schoolName": "云南财经",
		"province":   "云南",
	}

	//黑龙江职业学院     80562
	schoolMap[80562] = map[string]string{
		"schoolName": "黑龙江职业学院",
		"province":   "黑龙江",
	}

	//华中农业大学       78627
	schoolMap[78627] = map[string]string{
		"schoolName": "华中农业大学",
		"province":   "广东",
	}

	for _, schoolId := range schools {
		var nickName, headImgUrl, country, province, schoolName, deptName string
		var gender, age, grade, enrollmentYear, deptId int
		sql := fmt.Sprintf(`select nickName,headImgUrl,gender,age,grade,deptId,deptName,country,enrollmentYear 
from profiles where locate('14066660',phone) and length(headImgUrl)>0 ORDER BY RAND() limit 40`)
		rows, err1 := inst.Query(sql)
		if err1 != nil {
			err = err1
		}

		schoolName = schoolMap[schoolId]["schoolName"]
		province = schoolMap[schoolId]["province"]
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
						`, nickName, headImgUrl, gender, age, grade, schoolId, schoolName, deptId, deptName, country, province, enrollmentYear, rsp.Uin)

			_, err = inst.Exec(updateSql)
			if err != nil {
				err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
				log.Errorf(err.Error())
				return
			}
			log.Debugf("sql:%s", updateSql)

			//表示资料注册完成了，这时再更新通讯录里面的状态
			if len(nickName) > 0 {
				UpdateAddrBookInfo(rsp.Uin)

				//预先生成答题列表
				geneqids.Gene(rsp.Uin)

				//添加客服号
				var serviceAccountUin int64
				serviceAccountUin = 100001 //客服号
				sns.AddCustomServiceAccount(rsp.Uin, serviceAccountUin)

				//14066660301 客服号  102688
				sns.AddCustomServiceAccount(rsp.Uin, 102688)
			}

			//修改性别重新生成答题列表
			if gender >= 1 && gender <= 2  {
				//重新生成答题列表
				go geneqids.Gene(rsp.Uin)
			}

			//return
			phone++
		}

	}
	return

	//
	//update school info
}
