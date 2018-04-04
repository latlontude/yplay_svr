package submit

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"svr/st"
	"time"
)

type SubmitQueryListReq struct {
	Uin      int64  `schema:"uin"`
	Token    string `schema:"token"`
	Ver      int    `schema:"ver"`
	Type     int    `schema:"type"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

type SubmitQueryListRsp struct {
	Infos []*SubmitQInfo `json:"infos"`
}

type SubmitQInfo struct {
	SubmitId    int    `json:"submitId"`
	QId         int    `json:"qid"`
	QText       string `json:"qtext"`
	QIconUrl    string `json:"qiconUrl"`
	Status      int    `json:"status"`
	Desc        string `json:"desc"`
	VotedCnt    int    `json:"votedCnt"`
	NewVotedCnt int    `json:"newVotedCnt"`
	Flag        int    `json:"flag"` //是否是新上线 根据上次的时间戳来判断
}

func (this *SubmitQInfo) String() string {

	return fmt.Sprintf(`SubmitQInfo{SubmitId:%d, QId:%d, QText:%s, QIconUrl:%s, Status:%d, Desc:%s, VotedCnt:%d, Flag:%d}`,
		this.SubmitId, this.QId, this.QText, this.QIconUrl, this.Status, this.Desc, this.VotedCnt, this.Flag)
}

func doSubmitQueryList(req *SubmitQueryListReq, r *http.Request) (rsp *SubmitQueryListRsp, err error) {

	log.Debugf("uin %d, SubmitQueryListReq %+v", req.Uin, req)

	infos, err := SubmitQueryList(req.Uin, req.Type, req.PageNum, req.PageSize)
	if err != nil {
		log.Errorf("uin %d, SubmitQueryListRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &SubmitQueryListRsp{infos}

	log.Debugf("uin %d, SubmitQueryListRsp succ, %+v", req.Uin, rsp)

	return
}

func SubmitQueryList(uin int64, typ int, pageNum, pageSize int) (retInfos []*SubmitQInfo, err error) {

	log.Errorf("start SubmitQueryList")

	infos := make([]*SubmitQInfo, 0)
	retInfos = make([]*SubmitQInfo, 0)

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

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_SUBMIT_LAST_READ_ONLINE_TS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//全部从第一页开始计算
	if pageNum == 0 {
		pageNum = 1
	}

	if pageSize == 0 {
		pageSize = constant.DEFAULT_PAGE_SIZE
	}

	s := (pageNum - 1) * pageSize
	e := s + pageSize - 1

	sql := fmt.Sprintf(`select id, qid, qtext, qiconId, status, descr, mts from submitQuestions where uin = %d and status = %d order by mts desc`, uin, typ)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	qids := make([]int, 0)

	qidsMts := make(map[int]int)

	for rows.Next() {
		var info SubmitQInfo
		var qiconId int
		var mts int

		rows.Scan(&info.SubmitId, &info.QId, &info.QText, &qiconId, &info.Status, &info.Desc, &mts)
		info.QIconUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/qicon/%d.png", qiconId)

		qidsMts[info.QId] = mts

		infos = append(infos, &info)

		if info.Status == 1 {
			qids = append(qids, info.QId)
		}
	}

	if len(qids) == 0 {
		return
	}

	//已经上线的，判断是否新上线的问题列表
	//对比上次的拉取时间
	keyStr := fmt.Sprintf("%d", uin)
	valStr, err := app.Get(keyStr)

	if err != nil {

		//如果KEY不存在 则认为lastMsgId为0
		if e, ok := err.(*rest.APIError); ok {
			if e.Code == constant.E_REDIS_KEY_NO_EXIST {
				valStr = "0"
			} else {
				log.Errorf(err.Error())
				return
			}
		} else {
			log.Errorf(err.Error())
			return
		}
	}

	lastTs, err1 := strconv.Atoi(valStr)
	if err1 != nil {
		log.Errorf(err1.Error())
		lastTs = 0
	}

	//结束时要记录本次访问的时间
	defer func() {
		if err == nil {
			now := time.Now().Unix()
			app.Set(keyStr, fmt.Sprintf("%d", now))
		}
	}()

	qidsStr := ""
	for _, qid := range qids {
		qidsStr += fmt.Sprintf("%d,", qid)
	}
	qidsStr = qidsStr[:len(qidsStr)-1]

	//用户投稿的所有题的答题记录
	sql = fmt.Sprintf(`select qid, voteToUin, ts from voteRecords where qid in (%s) order by ts`, qidsStr)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	qidsTsMap := make(map[int]map[int64]int)
	for rows.Next() {
		var qid int
		var uin int64
		var ts int
		rows.Scan(&qid, &uin, &ts)

		if _, ok := qidsTsMap[qid]; !ok {
			qidsTsMap[qid] = make(map[int64]int)
		}
		qidsTsMap[qid][uin] = ts // ts 按升序排序，收到这道题投票的每一个用户的ts一定是最新的
	}

	//要全部查询出来 然后找同校同年级的
	sql = fmt.Sprintf(`select qid, voteToUin, count(id) as cnt from voteRecords where qid in (%s) group by qid, voteToUin`, qidsStr)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	//所有人的答题总数
	mqids := make(map[int]map[int64]int)
	for rows.Next() {

		var qid int
		var voteToUin int64
		var cnt int
		rows.Scan(&qid, &voteToUin, &cnt)

		if _, ok := mqids[qid]; !ok {
			mqids[qid] = make(map[int64]int)
		}

		mqids[qid][voteToUin] = cnt
	}
	//所有人的答题总数

	// 用户投稿的题新增的所有人答题记录
	sql = fmt.Sprintf(`select qid, voteToUin, count(id) as cnt from voteRecords where qid in (%s) and ts > %d group by qid, voteToUin`, qidsStr, lastTs)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	newQidsCntMap := make(map[int]map[int64]int)
	for rows.Next() {

		var qid int
		var voteToUin int64
		var cnt int
		rows.Scan(&qid, &voteToUin, &cnt)

		if _, ok := newQidsCntMap[qid]; !ok {
			newQidsCntMap[qid] = make(map[int64]int)
		}

		newQidsCntMap[qid][voteToUin] = cnt
	}

	uins := make([]int64, 0)
	for _, m := range mqids {
		for uid, _ := range m {
			uins = append(uins, uid)
		}
	}

	uins = append(uins, uin)

	res, err := st.BatchGetUserProfileInfo(uins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//同校同年级的答题总数
	mqidsFinal := make(map[int][]int)
	qidsUinMap := make(map[int][]int64) // 收到这题投票的同校同年级人的集合

	ui := res[uin]

	for qid, m := range mqids {
		cntTotal := 0
		newTotal := 0
		for uid, cnt := range m {
			if ui2, ok := res[uid]; ok {

				if ui2.SchoolId == ui.SchoolId && ui2.Grade == ui.Grade {

					/*	if ui.SchoolType == 3 && ui.DeptId != ui2.DeptId {
							continue // 用户学校为大学时，查找同校同学院同年级的用户
						}
					*/
					cntTotal += cnt

					if _, ok := newQidsCntMap[qid][uid]; ok {
						newTotal += newQidsCntMap[qid][uid]
					}

					qidsUinMap[qid] = append(qidsUinMap[qid], uid)
				}
			}
		}

		mqidsFinal[qid] = append(mqidsFinal[qid], cntTotal)
		mqidsFinal[qid] = append(mqidsFinal[qid], newTotal)
	}

	for i, info := range infos {
		if v, ok := mqidsFinal[info.QId]; ok {
			infos[i].VotedCnt = v[0]
			infos[i].NewVotedCnt = v[1]

		}

		//判断是否新上线的标志
		if mts, ok := qidsMts[info.QId]; ok {
			if mts > lastTs {
				infos[i].Flag = 1
			}
		}
	}

	qidsMaxTsMap := make(map[int]int) // 同校同年级的人收到该用户投稿的题目的投票的最新时间

	for _, info := range infos {
		if _, ok := qidsUinMap[info.QId]; ok {
			maxTs := 0
			for _, uin := range qidsUinMap[info.QId] {
				if qidsTsMap[info.QId][uin] > maxTs {
					maxTs = qidsTsMap[info.QId][uin]
				}
			}
			qidsMaxTsMap[info.QId] = maxTs
		} else {
			qidsMaxTsMap[info.QId] = 0
		}
	}

	log.Errorf("qidsMaxTsMap total:%d info :%+v", len(qidsMaxTsMap), qidsMaxTsMap)

	newHotCntQidMap := make(map[int][]int) //key 为新增热度数，value为相同新增热度数的用户投稿题目集合
	newCntSlice := make([]int, 0)

	for _, info := range infos {
		newHotCntQidMap[info.NewVotedCnt] = append(newHotCntQidMap[info.NewVotedCnt], info.QId)
	}

	log.Errorf("newHotCntQidMap : %+v", newHotCntQidMap)

	for key, _ := range newHotCntQidMap {
		newCntSlice = append(newCntSlice, key)
	}

	sort.Ints(newCntSlice[:]) // 按新增热度排序
	tmpInfos := make([]*SubmitQInfo, 0)

	for i := len(newCntSlice) - 1; i >= 0; i-- {

		tsQidMap := make(map[int][]int)
		tsSlice := make([]int, 0)

		for _, qid := range newHotCntQidMap[newCntSlice[i]] {
			ts := qidsMaxTsMap[qid]
			tsQidMap[ts] = append(tsQidMap[ts], qid)

		}

		for key, _ := range tsQidMap {

			tsSlice = append(tsSlice, key)
		}

		sort.Ints(tsSlice[:])
		log.Errorf("tsQidMap %+v", tsQidMap)
		log.Errorf("tsSlice total:%d, info:%+v", len(tsSlice), tsSlice)

		for j := len(tsSlice) - 1; j >= 0; j-- {
			for _, qid := range tsQidMap[tsSlice[j]] {
				for _, info := range infos {
					if info.QId == qid {
						tmpInfos = append(tmpInfos, info)
					}
				}
			}
		}

	}

	if len(tmpInfos) < s {
		return
	} else {
		start := s
		end := e
		if len(tmpInfos)-1 < e {
			end = len(tmpInfos) - 1
		}

		log.Errorf("total count:%d start:%d, end:%d", len(tmpInfos), start, end)

		for i := 0; i <= end-start; i++ {
			retInfos = append(retInfos, tmpInfos[start+i])
			log.Errorf("uin:%d, submitQid:%d, newly added count:%d, now total count:%d", uin, tmpInfos[start+i].QId, tmpInfos[start+i].NewVotedCnt, tmpInfos[start+i].VotedCnt)
		}
	}

	log.Errorf("end SubmitQueryList")
	return
}
