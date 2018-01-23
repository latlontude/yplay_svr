package vote

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"common/util"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"svr/cache"
	"svr/st"
	"time"
)

func GetNextQuestionAndOptionsByPreGene(uin int64, uuid int64) (qinfo *st.QuestionInfo, options []*st.OptionInfo2, index int, err error) {

	log.Debugf("GetNextQuestionAndOptions2 begin, uin %d, uuid %d", uin, uuid)

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_progress", uin)

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"qid", "qindex", "options", "voted", "cursor"}

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, HMGet rsp %+v", uin, valsStr)

	if len(valsStr) != len(fields) && len(valsStr) != 0 {
		err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
		log.Errorf(err.Error())
		return
	}

	lastQId := 0     //上一次的题目ID
	lastQIndex := 0  //上一次的答题编号 1~15
	voted := 1       //是否已经投票
	lastCursor := -1 //问题队列的上次扫描位置
	optionsStr := "" //上一次的选项列表

	//如果从来没有答题 则上一次题目设置为0 上答题一次索引为0
	if len(valsStr) != 0 {
		lastQId, _ = strconv.Atoi(valsStr["qid"])
		lastQIndex, _ = strconv.Atoi(valsStr["qindex"])
		voted, _ = strconv.Atoi(valsStr["voted"])
		optionsStr = valsStr["options"]

		lastCursor, _ = strconv.Atoi(valsStr["cursor"])
	}

	log.Debugf("uin %d, lastQId %d, lastQIndex %d, voted %d, cursor %d, optionsStr %s", uin, lastQId, lastQIndex, voted, lastCursor, optionsStr)

	//上次题目未回答
	if voted == 0 {

		log.Debugf("uin %d has unvoted question, optionStr %s", uin, optionsStr)

		err1 := json.Unmarshal([]byte(optionsStr), &options)
		if err1 != nil {
			log.Errorf(err1.Error())
		} else {
			//返回上次的选项和答题
			if q, ok := cache.QUESTIONS[lastQId]; ok {
				qinfo = q
				index = lastQIndex
				return
			}
		}
	}

	//找到下一题及新的游标
	newQId, newCursor, err := GetNextQIdByPreGeneCursor(uin, lastCursor)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	qinfo = cache.QUESTIONS[newQId]
	nextQGender := qinfo.OptionGender

	index = lastQIndex + 1
	if index > 15 {
		index = 1
	}

	log.Debugf("uin %d, genen newQId %d, newIndex %d, newGender %d, newCursor %d", uin, newQId, index, nextQGender, newCursor)

	//以下代码是选项计算逻辑

	//以后要优化最多拉取500个好友
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

	friendTs := make(map[int]int) //按加好友的时间排序

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

		friendTs[int(fi.Uin)] = fi.Ts

		friendUins = append(friendUins, fi.Uin)
	}

	log.Debugf("uin %d, boyCnt %d, girlCnt %d", uin, boyCnt, girlCnt)

	//获取好友的钻石列表
	gemCntMap, err := GetUinsGemCnt(friendUins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	// 加好友的时间(小时数) + 好友按钻石数
	// 最小的在最前
	nowTs := int(time.Now().Unix())
	for uid, v := range friendTs {
		tv := (nowTs - v) / 3600

		if tv < 0 {
			tv = 0
		}
		if v1, ok := gemCntMap[int64(uid)]; ok {
			tv += v1
		}
		friendTs[uid] = tv
	}

	uinsByAddFriendTime := make([]int64, 0)

	//最小的在最前
	pairs := util.SortMap1(friendTs)
	for _, p := range pairs {
		uinsByAddFriendTime = append(uinsByAddFriendTime, int64(p.Key))
	}

	log.Debugf("uin %d, gemCnt+TimeBeFriend orders %+v", uin, pairs)

	//获取该题目下我的好友的被投票的排序最多的在最前
	uinsVoteCntMap, err := st.GetUinsVotedCnt(newQId, friendUins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//按好友被投数目排序最多的在最前
	uinsByVote := make([]int64, 0)

	pairs = util.ReverseSortMap1(uinsVoteCntMap)
	for _, p := range pairs {
		uinsByVote = append(uinsByVote, int64(p.Key))
	}

	//最近三周的活跃度
	uinsByPVCnt := make([]int64, 0)

	uinsPVCntMap, err := GetUinsPVCnt(friendUins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//按PV排序最小的在最前
	pairs = util.SortMap1(uinsPVCntMap)
	for _, p := range pairs {
		uinsByPVCnt = append(uinsByPVCnt, int64(p.Key))
	}

	//预先生成好后续的换一换的信息
	prepared := ""

	//结束的时候，把信息写到redis
	defer func() {

		if err == nil {

			log.Debugf("uin %d, GetNextQuestionAndOptions2, in defer func, options %+v", uin, options)

			if len(options) != constant.ENUM_OPTION_BATCH_SIZE {
				err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
				log.Errorf(err.Error())
				return
			}

			randOrders := rand.Perm(constant.ENUM_OPTION_BATCH_SIZE)
			noptions := make([]*st.OptionInfo2, 0)

			for i, _ := range randOrders {
				noptions = append(noptions, options[randOrders[i]])
			}

			options = noptions

			//记录本次的选项集合信息,方便下次换一换时加快处理
			kvs := make(map[string]string)

			optionsTStr, _ := json.Marshal(options)
			optionsStr = string(optionsTStr)

			kvs["qid"] = fmt.Sprintf("%d", newQId)
			kvs["qindex"] = fmt.Sprintf("%d", index)
			kvs["options"] = optionsStr
			kvs["voted"] = "0"
			kvs["prepared"] = prepared
			kvs["preparedcursor"] = "3"
			kvs["cursor"] = fmt.Sprintf("%d", newCursor)

			err = app.HMSet(keyStr, kvs)
			if err != nil {
				log.Error(err.Error())
				return
			}
		}
	}()

	//好友人数小于4人, 当前好友 + 单项添加过的好友 + 通讯录好友 + 默认补充
	if len(friendUins) < 4 {

		//预先计算好的选项字符串
		for _, uid := range friendUins {
			prepared += fmt.Sprintf("%d:", uid)
		}

		if len(prepared) > 0 {
			prepared = prepared[:len(prepared)-1]
		} else {
			prepared = ""
		}

		log.Debugf("uin %d, friendUins(%d)<4,  prepared %s", uin, len(friendUins), prepared)

		combinOptions, err1 := GetOptionsByCombine(uin, uuid, friendUins, nextQGender, 4-len(friendUins))
		if err1 != nil {
			err = err1
			log.Errorf(err.Error())
			return
		}

		log.Debugf("uin %d, friendUins(%d)<4,  GetOptionsByCombine %+v", uin, len(friendUins), combinOptions)

		//我的好友数据
		for _, uid := range friendUins {

			option := &st.OptionInfo2{uid, friendInfos[uid].NickName, uinsVoteCntMap[int(uid)]}
			options = append(options, option)
		}

		uids := make([]int64, 0)
		//单向加好友或者通讯录或者明星
		for _, option := range combinOptions {
			options = append(options, option)

			if option.Uin != 0 {
				uids = append(uids, option.Uin)
			}
		}

		if len(uids) == 0 {
			return
		}

		//单向加好友或者通讯录好友在该题目下被选择的次数
		voteCntMap, _ := st.GetUinsVotedCnt(newQId, uids)

		for i, option := range options {

			if v, ok := voteCntMap[int(option.Uin)]; ok {
				options[i].BeSelCnt = v
			}
		}

		return
	}

	if nextQGender != 0 {

		//从中过滤出男性朋友或者女性朋友
		newUinsByVote := make([]int64, 0)
		newUinsByAddFriendTime := make([]int64, 0)
		newUinsByPVCnt := make([]int64, 0)

		for _, uin := range uinsByVote {
			if friendInfos[uin].Gender == nextQGender {
				newUinsByVote = append(newUinsByVote, uin)
			}
		}

		for _, uin := range uinsByAddFriendTime {
			if friendInfos[uin].Gender == nextQGender {
				newUinsByAddFriendTime = append(newUinsByAddFriendTime, uin)
			}
		}

		for _, uin := range uinsByPVCnt {
			if friendInfos[uin].Gender == nextQGender {
				newUinsByPVCnt = append(newUinsByPVCnt, uin)
			}
		}

		uinsByVote = newUinsByVote
		uinsByAddFriendTime = newUinsByAddFriendTime
		uinsByPVCnt = newUinsByPVCnt
	}

	//前面已经校验过不可能<4
	if len(uinsByVote) < constant.ENUM_OPTION_BATCH_SIZE || len(uinsByAddFriendTime) < constant.ENUM_OPTION_BATCH_SIZE || len(uinsByPVCnt) < constant.ENUM_OPTION_BATCH_SIZE {
		err = rest.NewAPIError(constant.E_VOTE_INFO_ERR, "vote info error")
		log.Errorf(err.Error())
		return
	}

	if len(uinsByVote) > 12 {
		uinsByVote = uinsByVote[:12]
	}

	if len(uinsByAddFriendTime) > 12 {
		uinsByAddFriendTime = uinsByAddFriendTime[:12]
	}

	if len(uinsByPVCnt) > 12 {
		uinsByPVCnt = uinsByPVCnt[:12]
	}

	log.Debugf("uin %d, friendUins %d, uinsByVote %+v,  uinsByAddFriendTime %+v, uinsByPVCnt %+v", uin, len(friendUins), uinsByVote, uinsByAddFriendTime, uinsByPVCnt)

	randomUins := friendUins
	if nextQGender == 1 {
		randomUins = boyUins
	} else if nextQGender == 2 {
		randomUins = girlUins
	}

	allOptionUins, err := PrepareAllOptionsUin2(uin, uinsByVote, uinsByAddFriendTime, uinsByPVCnt, randomUins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uid := range allOptionUins {
		prepared += fmt.Sprintf("%d:", uid)
	}
	prepared = prepared[:len(prepared)-1]

	log.Debugf("uin %d, friendUins(%d)>4,  prepared %s", uin, len(friendUins), prepared)

	selectedUins := allOptionUins[:4]

	for _, uid := range selectedUins {
		option := &st.OptionInfo2{uid, friendInfos[uid].NickName, uinsVoteCntMap[int(uid)]}
		options = append(options, option)
	}

	return
}

func GetNextQIdByPreGeneCursor(uin int64, lastCursor int) (qid, newCursor int, err error) {

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_qids", uin)

	total, err := app.ZCard(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if total == 0 {
		err = rest.NewAPIError(constant.E_PRE_GENE_QIDS_LIST_ERR, "qids list error")
		log.Errorf(err.Error())
		return
	}

	if lastCursor >= total {
		lastCursor = -1
	}

	newCursor = lastCursor + 1
	if newCursor >= total {
		newCursor = 0
	}

	vals, err := app.ZRange(keyStr, newCursor, newCursor)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if len(vals) < 0 {
		err = rest.NewAPIError(constant.E_PRE_GENE_QIDS_LIST_ERR, "qids list error")
		log.Errorf(err.Error())
		return
	}

	qid, _ = strconv.Atoi(vals[0])
	if qid == 0 {
		err = rest.NewAPIError(constant.E_PRE_GENE_QIDS_LIST_ERR, "qids list error")
		log.Errorf(err.Error())
		return
	}

	return
}
