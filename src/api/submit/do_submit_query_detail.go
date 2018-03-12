package submit

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"svr/st"
)

type SubmitQueryDetailReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	QId   int    `schema:"qid"`
}

type SubmitQueryDetailRsp struct {
	Total int            `json:"total"`
	Infos []*QidStatInfo `json:"infos"`
}

type QidStatInfo struct {
	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	Grade      int    `json:"grade"`
	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	DeptId     int    `json:"deptId"`
	DeptName   string `json:"deptName"`
	SchoolName string `json:"schoolName"`
	VotedCnt   int    `json:"votedCnt"`
}

func doSubmitQueryDetail(req *SubmitQueryDetailReq, r *http.Request) (rsp *SubmitQueryDetailRsp, err error) {

	log.Debugf("uin %d, SubmitQueryDetailReq %+v", req.Uin, req)

	total, infos, err := SubmitQueryDetail(req.Uin, req.QId)
	if err != nil {
		log.Errorf("uin %d, SubmitQueryDetailRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitQueryDetailRsp{total, infos}

	log.Debugf("uin %d, SubmitQueryDetailRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitQueryDetail(uin int64, qid int) (total int, infos []*QidStatInfo, err error) {

	infos = make([]*QidStatInfo, 0)

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid param")
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	//要全部查询出来 然后找同校同年级的
	sql := fmt.Sprintf(`select voteToUin, count(id) as cnt from voteRecords where qid = %d group by voteToUin order by cnt desc`, qid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	users := make([]int64, 0)

	users = append(users, uin)

	for rows.Next() {
		var info QidStatInfo
		rows.Scan(&info.Uin, &info.VotedCnt)
		infos = append(infos, &info)

		users = append(users, info.Uin)
	}

	res, err := st.BatchGetUserProfileInfo(users)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//该用户一定存在
	ui := res[uin]

	//只取同校同年级的用户
	ninfos := make([]*QidStatInfo, 0)

	for _, info := range infos {

		if v, ok := res[info.Uin]; ok {

			//同校同年级的前3名
			if v.SchoolId == ui.SchoolId && v.Grade == ui.Grade {
				/*   if ui.SchoolType == 3 && ui.DeptId != v.DeptId {
				           continue // 用户学校为大学时，查找同校同学院同年级的用户
				} */

				//总数也只是计算同校同年级的
				total += info.VotedCnt

				if len(ninfos) < 3 {

					info.NickName = v.NickName
					info.HeadImgUrl = v.HeadImgUrl
					info.Gender = v.Gender
					info.Grade = v.Grade
					info.SchoolId = v.SchoolId
					info.SchoolType = v.SchoolType
					info.SchoolName = v.SchoolName
					info.DeptId = v.DeptId
					info.DeptName = v.DeptName

					ninfos = append(ninfos, info)
				}
			}

		}
	}

	infos = ninfos

	return
}
