package st

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"encoding/json"
	"fmt"
	"strconv"
)

//返回给客户端查询的结构
type VoteRecord struct {
	VoteRecordId int64 `json:"voteRecordId"`

	Uin        int64  `json:"uin"`
	NickName   string `json:"nickName"`
	HeadImgUrl string `json:"headImgUrl"`
	Gender     int    `json:"gender"`
	Age        int    `json:"age"`
	Grade      int    `json:"grade"`
	SchoolId   int    `json:"schoolId"`
	SchoolType int    `json:"schoolType"`
	SchoolName string `json:"schoolName"`

	VoteToUin        int64  `json:"voteToUin"`
	VoteToNickName   string `json:"voteToNickName"`
	VoteToHeadImgUrl string `json:"voteToHeadImgUrl"`
	VoteToGender     int    `json:"voteToGender"`
	VoteToAge        int    `json:"voteToAge"`
	VoteToGrade      int    `json:"voteToGrade"`
	VoteToSchoolId   int    `json:"voteToSchoolId"`
	VoteToSchoolType int    `json:"voteToSchoolType"`
	VoteToSchoolName string `json:"voteToSchoolName"`

	QId      int    `json:"qid"`
	QText    string `json:"qtext"`
	QIconUrl string `json:"qiconUrl"`

	Options []*OptionInfo2 `json:"options"`
	Status  int            `json:"status"` //当前投票状态  初始 回复 回复的回复
	Ts      int            `json:"ts"`     //投票时间

	ImSessionId string `josn:"imSessionId"`
}

type UserBeSelCntInfo struct {
	Uin      int64 `json:"uin"`
	QId      int   `json:"qid"`
	BeSelCnt int   `json:"beSelCnt"`
}

func GetVoteRecordInfo(id int64) (info *VoteRecord, err error) {

	if id == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	info = &VoteRecord{}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select id, uin, qid, voteToUin, options, status, ts, imSessionId from voteRecords where id = %d`, id)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var options string
	find := false

	for rows.Next() {
		rows.Scan(
			&info.VoteRecordId,
			&info.Uin,
			&info.QId,
			&info.VoteToUin,
			&options,
			&info.Status,
			&info.Ts,
			&info.ImSessionId)

		find = true
		break
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "res not found")
		log.Error(err.Error())
		return
	}

	uis, err1 := BatchGetUserProfileInfo([]int64{info.Uin, info.VoteToUin})
	if err1 == nil {

		ui, ok := uis[info.Uin]
		if ok {
			info.NickName = ui.NickName
			info.HeadImgUrl = ui.HeadImgUrl
			info.Gender = ui.Gender

			info.Age = ui.Age
			info.Grade = ui.Grade
			info.SchoolId = ui.SchoolId
			info.SchoolType = ui.SchoolType
			info.SchoolName = ui.SchoolName
		}

		ui, ok = uis[info.VoteToUin]
		if ok {
			info.VoteToNickName = ui.NickName
			info.VoteToHeadImgUrl = ui.HeadImgUrl
			info.VoteToGender = ui.Gender

			info.VoteToAge = ui.Age
			info.VoteToGrade = ui.Grade
			info.VoteToSchoolId = ui.SchoolId
			info.VoteToSchoolType = ui.SchoolType
			info.VoteToSchoolName = ui.SchoolName
		}

	} else {
		log.Error(err1.Error())
	}

	sql = fmt.Sprintf(`select qtext, qiconUrl from questions2 where qid = %d`, info.QId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&info.QText, &info.QIconUrl)
	}

	err1 = json.Unmarshal([]byte(options), &info.Options)
	if err1 != nil {
		log.Error(err1.Error())
	}

	return
}

func GetVoteRecordInfoByImSessionId(imSessionId string) (info *VoteRecord, err error) {

	if len(imSessionId) == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	info = &VoteRecord{}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select id, uin, qid, voteToUin, options, status, ts, imSessionId from voteRecords where imSessionId = "%s"`, imSessionId)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var options string
	find := false

	for rows.Next() {
		rows.Scan(
			&info.VoteRecordId,
			&info.Uin,
			&info.QId,
			&info.VoteToUin,
			&options,
			&info.Status,
			&info.Ts,
			&info.ImSessionId)

		find = true
		break
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "res not found")
		log.Error(err.Error())
		return
	}

	uis, err1 := BatchGetUserProfileInfo([]int64{info.Uin, info.VoteToUin})
	if err1 == nil {

		ui, ok := uis[info.Uin]
		if ok {
			info.NickName = ui.NickName
			info.HeadImgUrl = ui.HeadImgUrl
			info.Gender = ui.Gender

			info.Age = ui.Age
			info.Grade = ui.Grade
			info.SchoolId = ui.SchoolId
			info.SchoolType = ui.SchoolType
			info.SchoolName = ui.SchoolName
		}

		ui, ok = uis[info.VoteToUin]
		if ok {
			info.VoteToNickName = ui.NickName
			info.VoteToHeadImgUrl = ui.HeadImgUrl
			info.VoteToGender = ui.Gender

			info.VoteToAge = ui.Age
			info.VoteToGrade = ui.Grade
			info.VoteToSchoolId = ui.SchoolId
			info.VoteToSchoolType = ui.SchoolType
			info.VoteToSchoolName = ui.SchoolName
		}

	} else {
		log.Error(err1.Error())
	}

	sql = fmt.Sprintf(`select qtext, qiconUrl from questions2 where qid = %d`, info.QId)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&info.QText, &info.QIconUrl)
		info.QIconUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/qicon/%s", info.QIconUrl)
	}

	err1 = json.Unmarshal([]byte(options), &info.Options)
	if err1 != nil {
		log.Error(err1.Error())
	}

	return
}

func GetUinsVotedCnt(qid int, uins []int64) (res map[int]int, err error) {

	res = make(map[int]int)

	if len(uins) == 0 {
		return
	}

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_QID_VOTED_CNT)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keys := make([]string, 0)

	for _, uin := range uins {
		keys = append(keys, fmt.Sprintf("%d_%d", uin, qid))
	}

	r, err := app.MGet(keys)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uin := range uins {

		res[int(uin)] = 0

		key := fmt.Sprintf("%d_%d", uin, qid)
		if v, ok := r[key]; ok {
			res[int(uin)], _ = strconv.Atoi(v)
		}
	}

	return
}

func IncrUserQidVotedCnt(uin int64, qid int) (cnt int, err error) {

	if uin == 0 || qid == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_QID_VOTED_CNT)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_%d", uin, qid)

	cnt, err = app.Incr(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}
