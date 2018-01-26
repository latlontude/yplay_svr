package geneqids

import (
	"common/constant"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"svr/st"
	"time"
)

var (
	LESSLT4_QIDS []*st.QuestionInfo //特制题库 好友人数<4人
	GENE_QIDS    []*st.QuestionInfo //通用题库
	BOY_QIDS     []*st.QuestionInfo //男性题库
	GIRL_QIDS    []*st.QuestionInfo //女性题库

	SUBMIT_NOT_LATEST_WEEK []*st.QuestionInfo //7天前的投稿题库
	SUBMIT_LATEST_WEEK     []*st.QuestionInfo //7天后的投稿题库

	ALL_UINS_SCHOOL_GRADE map[int64]*SchoolGradeInfo //所有的用户学校信息
	ALL_SUBMIT_QIDS_UINS  map[int]int64
)

type SchoolGradeInfo struct {
	SchoolId int `json:"schoolId"`
	Grade    int `json:"grade"`
}

type GeneQIdsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	User int64 `schema:"user"`
}

type GeneQIdsRsp struct {
	Total int `json:"total"`
}

func doGene(req *GeneQIdsReq, r *http.Request) (rsp *GeneQIdsRsp, err error) {

	log.Errorf("uin %d, GeneQIdsReq %+v", req.Uin, req.User, req)

	total, err := Gene(req.User)
	if err != nil {
		log.Errorf("uin %d, user %d, GeneQIdsRsp error %s", req.Uin, req.User, err.Error())
		return
	}

	rsp = &GeneQIdsRsp{total}

	log.Errorf("uin %d, user %d, GeneQIdsRsp succ, %+v", req.Uin, req.User, rsp)

	return
}

func Gene(uin int64) (total int, err error) {

	rand.Seed(time.Now().Unix())

	err = GetAllQIds()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//所有提交的问题ID
	err = GetAllSubmitQidsUin()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//所有人的年级信息 用于计算提交问题的年级信息
	err = GetAllUinsSchoolGradeInfo()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	qids, err := GeneQIds(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

    qids, err = optimizeQidsByUserAct(uin, qids)
      if err != nil {
		log.Errorf(err.Error())
		return
	}

	err = UpdateQIds(uin, qids)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	total = len(qids)

	return
}

func optimizeQidsByUserAct(uin int64, qids []int) (optimizedQids []int, err error){

    inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

     sql := fmt.Sprintf(`select qid, subTagId1, subTagId2, subTagId3 from questions2 `)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	qidsMap := make(map[int][]int)
	for rows.Next() {
       var qid, subTagId1, subTagId2, subTagId3 int 
       rows.Scan(&qid, &subTagId1, &subTagId2, &subTagId3)
       qidsMap[qid] = append(qidsMap[qid], subTagId1)
       qidsMap[qid] = append(qidsMap[qid], subTagId2)
       qidsMap[qid] = append(qidsMap[qid], subTagId3)
	 }


	sql = fmt.Sprintf(`select qid, act from actRecords where uin = %d  order by ts `, uin)

	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()


    
	skipMap := make(map[int]int)
	answerMap := make(map[int]int)

	for rows.Next() {

		var act int
		var qid  int
		rows.Scan(&qid, &act)
        
        subTagId1 := qidsMap[qid][0]
        subTagId2 := qidsMap[qid][1]
        subTagId3 := qidsMap[qid][2]

	    if act == 0 {
	    	if subTagId1 != 0 {
	    		 skipMap[subTagId1]++
            }

            if subTagId2 != 0 {        
	    	     skipMap[subTagId2]++
            }

            if subTagId3 != 0 {
	    		 skipMap[subTagId3]++
            }
	    }  else {

	    	if subTagId1 != 0 {
	    		 answerMap[subTagId1]++
            }

            if subTagId2 != 0 {
	    	     answerMap[subTagId2]++
            }

            if subTagId3 != 0 { 
	    		 answerMap[subTagId3]++ 
            }
	    }
	}
     
     subTagMap := make(map[int][]int)

     for _, qid := range qids {

        subTagId1 := qidsMap[qid][0]
        subTagId2 := qidsMap[qid][1]
        subTagId3 := qidsMap[qid][2]

	   if subTagId1 != 0 {
	   	  	subTagMap[subTagId1] = append(subTagMap[subTagId1], qid)
	   }

	   if subTagId2 != 0 {
	   	  	subTagMap[subTagId2] = append(subTagMap[subTagId2], qid)
	   }

	   if subTagId3 != 0 {
	   	  	subTagMap[subTagId3] = append(subTagMap[subTagId3], qid)
	   }
    }

    for subTagId := range subTagMap {

         skipCnt := skipMap[subTagId]
         answerCnt := answerMap[subTagId]

         if skipCnt > 2 {
         	 cnt := len(subTagMap[subTagId]) * (answerCnt / (skipCnt + answerCnt))
         	 if cnt > 0 {
         	   a := rand.Perm(len(subTagMap[subTagId])) //随机化
         	   i := 0
			   for _, idx := range a {
                i++ 
			    optimizedQids = append(optimizedQids, subTagMap[subTagId][idx])
                 
                 if i >= cnt {
                 	break
                 }
			    }
			
              }

            } else {

                optimizedQids = append(optimizedQids, subTagMap[subTagId]...)
            }
        }

        return

}

func GetAllSubmitQidsUin() (err error) {

	if len(ALL_SUBMIT_QIDS_UINS) > 0 {
		return
	}

	ALL_SUBMIT_QIDS_UINS = make(map[int]int64)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, qid from submitQuestions where status > 0`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uin int64
		var qid int

		rows.Scan(&uin, &qid)

		if qid > 0 && uin > 0 {
			ALL_SUBMIT_QIDS_UINS[qid] = uin
		}
	}

	return
}

func GetAllUinsSchoolGradeInfo() (err error) {

	if len(ALL_UINS_SCHOOL_GRADE) > 0 {
		return
	}

	ALL_UINS_SCHOOL_GRADE = make(map[int64]*SchoolGradeInfo)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, schoolId, grade from profiles`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uin int64
		var si SchoolGradeInfo

		rows.Scan(&uin, &si.SchoolId, &si.Grade)

		ALL_UINS_SCHOOL_GRADE[uin] = &si
	}

	return
}

func GetAllQIds() (err error) {

	//已经加载过
	if len(GENE_QIDS) > 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	GENE_QIDS = make([]*st.QuestionInfo, 0)
	BOY_QIDS = make([]*st.QuestionInfo, 0)
	GIRL_QIDS = make([]*st.QuestionInfo, 0)
	SUBMIT_NOT_LATEST_WEEK = make([]*st.QuestionInfo, 0)
	SUBMIT_LATEST_WEEK = make([]*st.QuestionInfo, 0)
	LESSLT4_QIDS = make([]*st.QuestionInfo, 0)

	sql := fmt.Sprintf(`select qid, qtext, qiconUrl, optionGender, replyGender, schoolType, dataSrc, status, tagId, tagName, subTagId1, subTagName1, subTagId2, subTagName2, subTagId3, subTagName3, ts from questions2 where status = 0 order by qid `)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	ts := int(time.Now().Unix())

	gap := 7 * 24 * 3600

	for rows.Next() {
		var info st.QuestionInfo
		rows.Scan(&info.QId, &info.QText, &info.QIconUrl, &info.OptionGender, &info.ReplyGender, &info.SchoolType, &info.DataSrc, &info.Status, &info.TagId, &info.TagName,
			&info.SubTagId1, &info.SubTagName1, &info.SubTagId2, &info.SubTagName2, &info.SubTagId3, &info.SubTagName3, &info.Ts)

		info.QIconUrl = fmt.Sprintf("http://yplay-1253229355.image.myqcloud.com/qicon/%s", info.QIconUrl)

		if info.DataSrc == 0 { //普通题库

			if info.OptionGender == 0 {
				GENE_QIDS = append(GENE_QIDS, &info)
			} else if info.OptionGender == 1 {
				BOY_QIDS = append(BOY_QIDS, &info)
			} else if info.OptionGender == 2 {
				GIRL_QIDS = append(GIRL_QIDS, &info)
			}

		} else if info.DataSrc == 1 { //特制题库

			LESSLT4_QIDS = append(LESSLT4_QIDS, &info)

		} else if info.DataSrc == 2 { //投稿题库

			if ts-info.Ts > gap {
				SUBMIT_NOT_LATEST_WEEK = append(SUBMIT_NOT_LATEST_WEEK, &info)
			} else {
				SUBMIT_LATEST_WEEK = append(SUBMIT_LATEST_WEEK, &info)
			}
		}
	}

	log.Errorf("GENE_QID_cnt %d, BOY_QID_cnt %d, GIRL_QID_cnt %d, LESSLT4_QIDS %d, SUBMIT_NOT_LATEST_WEEK %d, SUBMIT_LATEST_WEEK %d",
		len(GENE_QIDS), len(BOY_QIDS), len(GIRL_QIDS), len(LESSLT4_QIDS), len(SUBMIT_LATEST_WEEK), len(SUBMIT_NOT_LATEST_WEEK))

	return
}

func GeneQIds(uin int64) (qids []int, err error) {

	//先查找已经回答的题目
	answneredIds, err := GetLastAnswneredQIds(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("uin %d, AnswneredQIds cnt %d", uin, len(answneredIds))

	//重新生成的列表 已经随机化
	preQIds, err := PreGeneUserQIds(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("uin %d, PreGeneUserQIds cnt %d", uin, len(preQIds))

	qids = make([]int, 0)

	//在已经回答过的题目里面 并且在新生成的列表里面
	join_qids := make([]int, 0)

	//已经回答的放到后面 未回答的放在前面
	for _, qid := range preQIds {

		find := false

		for _, t := range answneredIds {

			if qid == t {
				find = true
				join_qids = append(join_qids, t)
				break
			}
		}

		if !find {
			qids = append(qids, qid)
		}
	}

	log.Errorf("uin %d, join_qids cnt %d", uin, len(join_qids))

	//已经回答的放到前面
	for _, t := range join_qids {
		qids = append(qids, t)
	}

	log.Errorf("uin %d, last gene qids cnt %d", uin, len(qids))

	return
}

func rand_order_qids(qids []int) (rand_qids []int) {
	rand_qids = make([]int, 0)

	if len(qids) == 0 {
		return
	}

	curPos := 0

	//100道题内随机化
	for {
		endPos := curPos + 100
		if endPos > len(qids) {
			endPos = len(qids)
		}

		//结束
		if endPos <= curPos {
			break
		}

		curQIds := qids[curPos:endPos]

		//随机化100题
		a := rand.Perm(endPos - curPos)
		for _, t := range a {
			rand_qids = append(rand_qids, curQIds[t])
		}

		curPos = endPos
	}

	return
}

func GetLastAnswneredQIds(uin int64) (qids []int, err error) {

	qids = make([]int, 0)

	if uin == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"cursor"}

	keyStr := fmt.Sprintf("%d_progress", uin)

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, GetLastAnswneredQIds HMGet rsp %+v", uin, valsStr)

	if len(valsStr) != len(fields) && len(valsStr) != 0 {
		err = rest.NewAPIError(constant.E_PRE_GENE_QIDS_PROGRESS_ERR, "pre gene progress info error")
		log.Errorf(err.Error())
		return
	}

	lastCursor := -1 //队列的上次扫描位置

	//如果从来没有答题 则上一次题目设置为0  上答题一次索引为0
	if len(valsStr) != 0 {
		lastCursor, _ = strconv.Atoi(valsStr["cursor"])
	}

	//还从未答过本轮题目 则认为已经回答的题目为空
	if lastCursor < 0 {
		return
	}

	keyStr2 := fmt.Sprintf("%d_qids", uin)

	vals, err := app.ZRange(keyStr2, 0, lastCursor)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, val := range vals {
		qid, _ := strconv.Atoi(val)
		if qid > 0 {
			qids = append(qids, qid)
		}
	}

	return
}

func UpdateQIds(uin int64, qids []int) (err error) {

	if uin == 0 || len(qids) == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"cursor", "qid", "qindex", "options", "voted", "prepared", "preparedCursor"}

	keyStr := fmt.Sprintf("%d_progress", uin)

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, UpdateQIds HMGet rsp %+v", uin, valsStr)

	if len(valsStr) != len(fields) && len(valsStr) != 0 {
		err = rest.NewAPIError(constant.E_PRE_GENE_QIDS_PROGRESS_ERR, "pre gene qids progress info error")
		log.Errorf(err.Error())
		return
	}

	cursor := "-1"
	qid := "0" //上一次的题目ID
	qindex := "0"
	options := ""
	voted := "1"
	prepared := ""
	preparedcursor := "-1"

	//如果从来没有答题 则上一次题目设置为0  上答题一次索引为0
	if len(valsStr) != 0 {
		qid = valsStr["qid"]
		qindex = valsStr["qindex"]
		options = valsStr["options"]
		voted = valsStr["voted"]
		prepared = valsStr["prepared"]
		preparedcursor = valsStr["preparedcursor"]
	}

	//清理原有的
	keyStr2 := fmt.Sprintf("%d_qids", uin)

	err = app.Del(keyStr2)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//新生成的
	for i, qid := range qids {
		err = app.ZAdd(keyStr2, int64(i), fmt.Sprintf("%d", qid))
		if err != nil {
			log.Errorf(err.Error())
		}
	}

	//重置进度
	valsMap := make(map[string]string)
	valsMap["cursor"] = cursor                 //生成的题目的游标
	valsMap["qid"] = qid                       //题目ID
	valsMap["qindex"] = qindex                 //客户端索引编号1~15
	valsMap["options"] = options               //保存未回答的选项
	valsMap["voted"] = voted                   //是否已经回答
	valsMap["prepared"] = prepared             //预先生成的选项
	valsMap["preparedcursor"] = preparedcursor //选项游标

	log.Debugf("uin %d, UpdateQIds HMSet %+v", uin, valsMap)

	err = app.HMSet(keyStr, valsMap)
	if err != nil {
		log.Errorf(err.Error())
	}

	return
}

func PreGeneUserQIds(uin int64) (qids []int, err error) {

	log.Errorf("uin %d, begin PreGeneQIds....", uin)

	qids = make([]int, 0)

	defer func() {

		if err == nil {
			log.Errorf("uin %d, end PreGeneQIds, qids cnt %d", uin, len(qids))
		}

	}()

	if uin == 0 {
		return
	}

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("uin %d, PreGeneUserQIds nickName %s, gender %d, schoolId %d, schoolType %d, schoolName %s, grade %d", uin, ui.NickName, ui.Gender, ui.SchoolId, ui.SchoolType, ui.SchoolName, ui.Grade)

	//question里面的schoolType 1/2/4  1+2=3 1+4=5 2+4=6 1+2+4=7
	if ui.SchoolType == 3 {
		ui.SchoolType = 4
	}

	friendInfos, err := st.GetAllMyFriends(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	boyCnt := 0
	girlCnt := 0
	friendUins := make([]int64, 0)

	boyUins := make([]int64, 0)
	girlUins := make([]int64, 0)

	for _, fi := range friendInfos {

		//如果没有性别或者昵称为空，则过滤掉
		if fi.Gender == 0 || len(fi.NickName) == 0 {
			continue
		}

		//统计男/女生的数目
		if fi.Gender == 1 {
			boyCnt += 1
			boyUins = append(boyUins, fi.Uin)
		} else {
			girlCnt += 1
			girlUins = append(girlUins, fi.Uin)
		}

		friendUins = append(friendUins, fi.Uin)
	}

	log.Errorf("uin %d, PreGeneUserQIds boyCnt %d, girlCnt %d, friendCnt %d", uin, boyCnt, girlCnt, len(friendUins))

	//好友人数<4 特制题库+通用题库
	//通用题库还要筛选学校
	if len(friendUins) < 4 {

		//特制题库
		//随机化
		if len(LESSLT4_QIDS) > 0 {
			a := rand.Perm(len(LESSLT4_QIDS)) //随机化
			for _, idx := range a {

				qinfo := LESSLT4_QIDS[idx]

				if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
					qids = append(qids, qinfo.QId)
				}
			}
		}

		//最近7天的投稿题库
		//性别->通用/学校->匹配
		//随机化

		if len(SUBMIT_LATEST_WEEK) > 0 {
			a := rand.Perm(len(SUBMIT_LATEST_WEEK))
			for _, idx := range a {
				qinfo := SUBMIT_LATEST_WEEK[idx]
				if qinfo.OptionGender == 0 {
					if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

						//获取投稿人的UIN
						if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
							continue
						}
						qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

						//获取投稿人的学校信息
						if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
							continue
						}
						si := ALL_UINS_SCHOOL_GRADE[qidUin]

						if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

							//同校同年级的才加入
							if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
								qids = append(qids, qinfo.QId)
							}
						}
					}
				}
			}
		}

		//普通题库
		//随机化
		order_qids := make([]int, 0)
		for _, qinfo := range GENE_QIDS {
			if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

				if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
					order_qids = append(order_qids, qinfo.QId)
				}
			}
		}

		//远于7天的投稿题库
		//性别->通用/学校->匹配
		for _, qinfo := range SUBMIT_NOT_LATEST_WEEK {
			if qinfo.OptionGender == 0 {
				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					//获取投稿人的UIN
					if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
						continue
					}
					qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

					//获取投稿人的学校信息
					if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
						continue
					}
					si := ALL_UINS_SCHOOL_GRADE[qidUin]

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

						//同校同年级的才加入
						if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
							order_qids = append(order_qids, qinfo.QId)
						}
					}
				}
			}
		}

		//普通题库随机化
		rand_qids := rand_order_qids(order_qids)
		for _, qid := range rand_qids {
			qids = append(qids, qid)
		}

		return
	}

	//boyUins >= 4 and girlUins >= 4  通用 + 男性题目 + 女性题目 +  学校筛选
	if len(boyUins) >= 4 && len(girlUins) >= 4 {

		//最近7天的投稿题库 性别 + 学校
		//随机化
		if len(SUBMIT_LATEST_WEEK) > 0 {
			a := rand.Perm(len(SUBMIT_LATEST_WEEK))
			for _, idx := range a {

				qinfo := SUBMIT_LATEST_WEEK[idx]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					//获取投稿人的UIN
					if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
						continue
					}
					qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

					//获取投稿人的学校信息
					if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
						continue
					}
					si := ALL_UINS_SCHOOL_GRADE[qidUin]

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						//同校同年级的才加入
						if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
							qids = append(qids, qinfo.QId)
						}

					}
				}
			}
		}

		//普通题库
		i := 0
		order_qids := make([]int, 0)

		for {

			hasMore := false

			if i < len(GENE_QIDS) {
				hasMore = true

				qinfo := GENE_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			if i < len(BOY_QIDS) {
				hasMore = true

				qinfo := BOY_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			if i < len(GIRL_QIDS) {
				hasMore = true
				qinfo := GIRL_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			i++

			if !hasMore {
				break
			}
		}

		//远于7天的投稿题库
		for _, qinfo := range SUBMIT_NOT_LATEST_WEEK {
			if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

				if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

					//获取投稿人的UIN
					if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
						continue
					}
					qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

					//获取投稿人的学校信息
					if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
						continue
					}
					si := ALL_UINS_SCHOOL_GRADE[qidUin]

					//同校同年级的才加入
					if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}
		}

		//普通题库随机化
		rand_qids := rand_order_qids(order_qids)

		for _, qid := range rand_qids {
			qids = append(qids, qid)
		}

		return
	}

	//boyUins >= 4 and girlUins < 4   通用 + 男性题目 + 学校筛选
	if len(boyUins) >= 4 && len(girlUins) < 4 {
		//最近7天的投稿题库

		if len(SUBMIT_LATEST_WEEK) > 0 {

			a := rand.Perm(len(SUBMIT_LATEST_WEEK))

			for _, idx := range a {

				qinfo := SUBMIT_LATEST_WEEK[idx]
				if qinfo.OptionGender == 0 || qinfo.OptionGender == 1 {

					if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

						//获取投稿人的UIN
						if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
							continue
						}
						qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

						//获取投稿人的学校信息
						if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
							continue
						}
						si := ALL_UINS_SCHOOL_GRADE[qidUin]

						if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

							//同校同年级的才加入
							if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
								qids = append(qids, qinfo.QId)
							}

						}
					}
				}
			}
		}

		i := 0
		order_qids := make([]int, 0)

		for {

			hasMore := false

			if i < len(GENE_QIDS) {
				hasMore = true
				qinfo := GENE_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			if i < len(BOY_QIDS) {
				hasMore = true
				qinfo := BOY_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			i++

			if !hasMore {
				break
			}
		}

		//7天之前的的投稿题库
		for _, qinfo := range SUBMIT_NOT_LATEST_WEEK {
			if qinfo.OptionGender == 0 || qinfo.OptionGender == 1 {

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					//获取投稿人的UIN
					if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
						continue
					}
					qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

					//获取投稿人的学校信息
					if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
						continue
					}
					si := ALL_UINS_SCHOOL_GRADE[qidUin]

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

						//同校同年级的才加入
						if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
							order_qids = append(order_qids, qinfo.QId)
						}

					}
				}
			}
		}

		//普通题库随机化
		rand_qids := rand_order_qids(order_qids)
		for _, qid := range rand_qids {
			qids = append(qids, qid)
		}

		return
	}

	//boyUins < 4  and girlUins >= 4  通用 + 女性题目 + 学校筛选
	if len(boyUins) < 4 && len(girlUins) >= 4 {

		//7天内的的投稿题库
		if len(SUBMIT_LATEST_WEEK) > 0 {
			a := rand.Perm(len(SUBMIT_LATEST_WEEK))
			for _, idx := range a {

				qinfo := SUBMIT_LATEST_WEEK[idx]

				if qinfo.OptionGender == 0 || qinfo.OptionGender == 2 {
					if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

						if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

							//获取投稿人的UIN
							if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
								continue
							}
							qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

							//获取投稿人的学校信息
							if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
								continue
							}
							si := ALL_UINS_SCHOOL_GRADE[qidUin]

							//同校同年级的才加入
							if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
								qids = append(qids, qinfo.QId)
							}

						}
					}
				}
			}
		}

		i := 0
		order_qids := make([]int, 0)

		for {

			hasMore := false

			if i < len(GENE_QIDS) {
				hasMore = true
				qinfo := GENE_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			if i < len(GIRL_QIDS) {
				hasMore = true
				qinfo := GIRL_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			i++

			if !hasMore {
				break
			}
		}

		//7天前的的投稿题库
		for _, qinfo := range SUBMIT_NOT_LATEST_WEEK {
			if qinfo.OptionGender == 0 || qinfo.OptionGender == 2 {

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

						//获取投稿人的UIN
						if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
							continue
						}
						qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

						//获取投稿人的学校信息
						if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
							continue
						}
						si := ALL_UINS_SCHOOL_GRADE[qidUin]

						//同校同年级的才加入
						if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
							order_qids = append(order_qids, qinfo.QId)
						}

					}
				}
			}
		}

		//普通题库随机化
		rand_qids := rand_order_qids(order_qids)
		for _, qid := range rand_qids {
			qids = append(qids, qid)
		}

		return

	}

	//boyUins < 4  and girlUins < 4  通用 + 学校筛选
	if len(boyUins) < 4 && len(girlUins) < 4 {

		//7天内的的投稿题库
		if len(SUBMIT_LATEST_WEEK) > 0 {
			a := rand.Perm(len(SUBMIT_LATEST_WEEK))
			for _, idx := range a {
				qinfo := SUBMIT_LATEST_WEEK[idx]
				if qinfo.OptionGender == 0 {
					if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

						if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

							//获取投稿人的UIN
							if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
								continue
							}
							qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

							//获取投稿人的学校信息
							if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
								continue
							}
							si := ALL_UINS_SCHOOL_GRADE[qidUin]

							//同校同年级的才加入
							if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
								qids = append(qids, qinfo.QId)
							}

						}
					}
				}
			}
		}

		i := 0
		order_qids := make([]int, 0)
		for {

			hasMore := false

			if i < len(GENE_QIDS) {
				hasMore = true
				qinfo := GENE_QIDS[i]

				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {
						order_qids = append(order_qids, qinfo.QId)
					}
				}
			}

			i++
			if !hasMore {
				break
			}
		}

		//7天内的的投稿题库
		for _, qinfo := range SUBMIT_NOT_LATEST_WEEK {
			if qinfo.OptionGender == 0 {
				if qinfo.SchoolType == 0 || (qinfo.SchoolType&ui.SchoolType) > 0 {

					if qinfo.ReplyGender == 0 || qinfo.ReplyGender == ui.Gender {

						//获取投稿人的UIN
						if _, ok := ALL_SUBMIT_QIDS_UINS[qinfo.QId]; !ok {
							continue
						}
						qidUin := ALL_SUBMIT_QIDS_UINS[qinfo.QId]

						//获取投稿人的学校信息
						if _, ok := ALL_UINS_SCHOOL_GRADE[qidUin]; !ok {
							continue
						}
						si := ALL_UINS_SCHOOL_GRADE[qidUin]

						//同校同年级的才加入
						if si.SchoolId == ui.SchoolId && si.Grade == ui.Grade {
							order_qids = append(order_qids, qinfo.QId)
						}
					}
				}
			}
		}

		//普通题库随机化
		rand_qids := rand_order_qids(order_qids)
		for _, qid := range rand_qids {
			qids = append(qids, qid)
		}

		return
	}

	return
}
