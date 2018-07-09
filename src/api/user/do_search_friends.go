package user

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"svr/st"
)
type uInfo []*st.UserProfileInfo

func (u uInfo) Len() int {
	return len(u)
}

// < 升序 0 1 2
// > 降序 2 1 0

func (u uInfo) Less(i, j int) bool {
	return u[i].Src < u[j].Src
}

func (u uInfo) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}


type SearchInfoReq struct{
	Uin         int64  `schema:"uin"`
	Token       string `schema:"token"`
	Ver         int    `schema:"ver"`

	City        string  `schema:"city"`

	Gender      int     `json:"gender"`
	SchoolId    int     `json:"schoolId"`
	SchoolName  string  `schema:"schoolName"`
	Grade       int     `json:"grade"`           //年级
	DeptId      int     `json:"deptId"`   //大学的学院信息
	DeptName    string  `json:"deptName"` //大学的学院信息

	HomeTown    string  `schema:"hometown"`  //家乡

	PageNum     int     `schema:"pageNum"`
	PageSize    int     `schema:"pageSize"`

	SearchType    int  `schema:"searchType"`      //按什么搜索  0:同城   1:同校   2:男   3:女   4:同学院  5:同年级 6:同乡
}


type SearchInfoRsp struct {
	Info []*st.UserProfileInfo `json:"info"`        //用户资料
	TotalCnt int               `json:"totalCnt"`    //推荐好友总数
}






/**
	找同学 查找和学校在同一城市的用户
 */

func doSearchFriends(req *SearchInfoReq, r *http.Request) (rsp *SearchInfoRsp, err error){
	log.Errorf("uin %d, SearchInfoReq %+v", req.Uin, req)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}


	//先获取自己的userinfo
	myInfo ,err :=  st.GetUserProfileInfo(req.Uin)

	var info []*st.UserProfileInfo
	var totalCnt int


	if req.PageNum == 0 {
		req.PageNum = 1
	}

	if req.PageSize == 0 {
		req.PageSize = constant.DEFAULT_PAGE_SIZE
	}


	switch req.SearchType {
		case 0:     //同城
			info,totalCnt ,err = findSameCityFriends(myInfo,inst, req.City, req.PageNum,req.PageSize)
		case 1:     //同学校
			info,totalCnt ,err = findSameSchoolFriends(myInfo,inst,req.SchoolId, req.SchoolName,req.PageNum,req.PageSize)
		case 2,3:   //男 女
			info,totalCnt ,err = findSameGenderFriends(myInfo,inst,req.Gender,req.PageNum,req.PageSize)
		case 4:     //同学院
			info,totalCnt ,err = findSameDepartFriends(myInfo,inst,req.DeptName,req.PageNum,req.PageSize)
		case 5:     //同年级
			info,totalCnt ,err = findSameGradeFriends(myInfo,inst,req.Grade,req.PageNum,req.PageSize)
		case 6:     //同乡
			info,totalCnt ,err = findSameHometownFriends(myInfo,inst,req.HomeTown,req.PageNum,req.PageSize)
	}

	if err != nil {
		log.Errorf("uin %d, SearchType:%d , SearchInfoRsp error %s", req.Uin, req.SearchType,err.Error())
		return
	}

	rsp = &SearchInfoRsp{info,totalCnt}

	log.Errorf("uin %d, SearchInfoRsp succ, %+v", req.Uin, rsp)

	return
}

/**
	同乡 type = 6
 */
func findSameHometownFriends(myInfo *st.UserProfileInfo ,inst *sql.DB ,hometown string,pageNum int, pageSize int) (info []*st.UserProfileInfo , totalCnt int , err error){


	if len(hometown) == 0 {
		log.Errorf("hometown is empty ,please check params")
		return
	}



	sql1 := fmt.Sprintf(`select count(*) from profiles where hometown = '%s' and uin != %d`,hometown,myInfo.Uin)
	totalCnt ,err = searchTotalCnt(inst,sql1)

	//数据库查找范围
	start := (pageNum -1) * pageSize
	end := start + pageSize
	if end > totalCnt {
		end = totalCnt
	}


	//模糊匹配  深圳 深圳市
	sql2 := fmt.Sprintf(`select uin from profiles where hometown = '%s'  and uin != %d limit  %d , %d `,hometown,myInfo.Uin,start,end)
	info , err = search(myInfo,inst,sql2)

	return
}

/**
	同城
 */
func findSameCityFriends(myInfo *st.UserProfileInfo ,inst *sql.DB ,city string,pageNum int, pageSize int) (info []*st.UserProfileInfo , totalCnt int , err error){


	if len(city) == 0 {
		log.Errorf("city is empty ,please check params")
		return
	}

	cityEnd := "市"
	s := strings.Split(city,cityEnd)
	cityName := s[0]

	sql1 := fmt.Sprintf(`select count(*) from profiles where city = '%s' or city = '%s' and uin != %d`,cityName,cityName+cityEnd,myInfo.Uin)
	totalCnt ,err = searchTotalCnt(inst,sql1)

	//数据库查找范围
	start := (pageNum -1) * pageSize
	end := start + pageSize
	if end > totalCnt {
		end = totalCnt
	}


	//模糊匹配  深圳 深圳市
	sql2 := fmt.Sprintf(`select uin from profiles where city = '%s' or city = '%s'  and uin != %d limit  %d , %d `,cityName,cityName+cityEnd,myInfo.Uin,start,end)
	info , err = search(myInfo,inst,sql2)

	return
}


/**
	同校
 */
func findSameSchoolFriends(myInfo *st.UserProfileInfo ,inst *sql.DB ,schoolId int, schoolName string,pageNum int, pageSize int) (info []*st.UserProfileInfo , totalCnt int , err error){


	sql1 := fmt.Sprintf(`select count(*) from profiles where schoolId = %d and schoolName = '%s' and uin != %d`,schoolId,schoolName,myInfo.Uin)

	totalCnt ,err = searchTotalCnt(inst,sql1)

	//数据库查找范围
	start := (pageNum -1) * pageSize
	end := start + pageSize

	if end > totalCnt {
		end = totalCnt
	}


	sql2 := fmt.Sprintf(`select uin from profiles where schoolId = %d and schoolName = '%s' and uin != %d limit  %d , %d `,schoolId,schoolName,myInfo.Uin,start,end)

	info ,err = search(myInfo,inst,sql2)

	return
}


/**
	性别
 */
func findSameGenderFriends(myInfo *st.UserProfileInfo ,inst *sql.DB ,gender int,pageNum int, pageSize int) (info []*st.UserProfileInfo , totalCnt int , err error){

	sql1 := fmt.Sprintf(`select count(*) from profiles where gender = %d and uin != %d `,gender,myInfo.Uin)
	totalCnt ,err = searchTotalCnt(inst,sql1)

	//数据库查找范围
	start := (pageNum -1) * pageSize
	end := start + pageSize
	if end > totalCnt {
		end = totalCnt
	}

	sql2 := fmt.Sprintf(`select uin from profiles where gender = %d and uin != %d limit  %d , %d `,gender,myInfo.Uin,start,end)

	info ,err = search(myInfo,inst,sql2)

	return
}


/**
	同学院
 */
func findSameDepartFriends(myInfo *st.UserProfileInfo ,inst *sql.DB ,deptName string,pageNum int, pageSize int) (info []*st.UserProfileInfo , totalCnt int , err error){

	sql1 := fmt.Sprintf(`select count(*) from profiles where deptName = '%s' and uin != %d `,deptName,myInfo.Uin)
	totalCnt ,err = searchTotalCnt(inst,sql1)

	//数据库查找范围
	start := (pageNum -1) * pageSize
	end := start + pageSize
	if end > totalCnt {
		end = totalCnt
	}

	sql2 := fmt.Sprintf(`select uin from profiles where deptName = '%s' and uin != %d limit  %d , %d `,deptName,myInfo.Uin,start,end)

	info ,err = search(myInfo,inst,sql2)

	return
}



/**
	同年级
 */
func findSameGradeFriends(myInfo *st.UserProfileInfo ,inst *sql.DB ,grade int ,pageNum int, pageSize int) (info []*st.UserProfileInfo , totalCnt int , err error){

	sql1 := fmt.Sprintf(`select count(*) from profiles where grade = %d and uin != %d`,grade,myInfo.Uin)

	totalCnt ,err = searchTotalCnt(inst,sql1)

	//数据库查找范围
	start := (pageNum -1) * pageSize
	end := start + pageSize
	if end > totalCnt {
		end = totalCnt
	}

	sql2 := fmt.Sprintf(`select uin from profiles where grade = %d and uin != %d limit  %d , %d `,grade,myInfo.Uin,start,end)

	info ,err = search(myInfo,inst,sql2)

	return
}




func searchTotalCnt(inst *sql.DB , sql string)( totalCnt int , err error){


	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&totalCnt)
	}

	return
}


func search(myInfo *st.UserProfileInfo ,inst *sql.DB , sql string) (info []*st.UserProfileInfo , err error){

	info = make([]*st.UserProfileInfo, 0)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}


	log.Debugf("myInfo:%v",myInfo)

	for rows.Next() {
		var uid int64
		rows.Scan(&uid)


		ui, err1 := st.GetUserProfileInfo(uid)

		if ui.SchoolId == myInfo.SchoolId && ui.SchoolName == myInfo.SchoolName {       //同校
			ui.Src = 0
		}else if ui.City == myInfo.City {                                               //同城
			ui.Src = 1
		}else if ui.Province == myInfo.Province {
			ui.Src = 3                                                                  //同省
		}else {
			ui.Src = 4                                                                  //其他
		}

		if err1 != nil {
			log.Error(err1.Error())
			continue
		}
		info = append(info, ui)
	}


	//排序
	sort.Sort(uInfo(info))
	return
}



