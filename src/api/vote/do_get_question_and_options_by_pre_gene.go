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
	lastQIndex := 0  //上一次的答题编号
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
	if index > constant.ENUM_QUESTION_BATCH_SIZE {
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
	noNickNameFriendsUins := make([]int64, 0)

	boyUins := make([]int64, 0)
	girlUins := make([]int64, 0)

	friendTs := make(map[int]int) //按加好友的时间排序

	for _, fi := range friendInfos {

		//如果没有性别或者昵称为空，则过滤掉
		if fi.Gender == 0 || len(fi.NickName) == 0 {
			noNickNameFriendsUins = append(noNickNameFriendsUins, fi.Uin)
			continue
		}

		if fi.Uin == 100001 {
			continue // 客服号不出现在选项里
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

	log.Debugf("uin %d, boyCnt %d, girlCnt %d, noNickNameFriendsCnt:%d (%+v)", uin, boyCnt, girlCnt, len(noNickNameFriendsUins), noNickNameFriendsUins)

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

			//统计每一轮的选项UIN出现的次数
			//如果是最后一轮 需要清理掉之前的统计数据
			if index == constant.ENUM_QUESTION_BATCH_SIZE {
				go ClearOptionUinsLookedCnt(uin)
			} else {
				go AddOptionUinsLookedCnt(uin, options)
			}
		}
	}()

	if nextQGender != 0 {

		//从中过滤出男性朋友或者女性朋友
		newUinsByVote := make([]int64, 0)
		newUinsByAddFriendTime := make([]int64, 0)
		newUinsByPVCnt := make([]int64, 0)

		for _, uid := range uinsByVote {
			if friendInfos[uid].Gender == nextQGender {
				newUinsByVote = append(newUinsByVote, uid)
			}
		}

		for _, uid := range uinsByAddFriendTime {
			if friendInfos[uid].Gender == nextQGender {
				newUinsByAddFriendTime = append(newUinsByAddFriendTime, uid)
			}
		}

		for _, uid := range uinsByPVCnt {
			if friendInfos[uid].Gender == nextQGender {
				newUinsByPVCnt = append(newUinsByPVCnt, uid)
			}
		}

		uinsByVote = newUinsByVote
		uinsByAddFriendTime = newUinsByAddFriendTime
		uinsByPVCnt = newUinsByPVCnt
	}

	randomUins := friendUins
	if nextQGender == 1 {
		randomUins = boyUins
	} else if nextQGender == 2 {
		randomUins = girlUins
	}

	// 为用户准备好题库后，用户增删过好友 导致一些题目已经不适合ta, 需要为该题目重新准备候选人
	if len(uinsByVote) < constant.ENUM_OPTION_BATCH_SIZE || len(uinsByAddFriendTime) < constant.ENUM_OPTION_BATCH_SIZE || len(uinsByPVCnt) < constant.ENUM_OPTION_BATCH_SIZE {
		log.Debugf(" start prepare candidates list")
		log.Debugf("uinsByvote:%+v uinsByAddFriendTime:%+v  uinsByPVCnt:%+v", uinsByVote, uinsByAddFriendTime, uinsByPVCnt)

		cnt := len(uinsByVote)
		if len(uinsByAddFriendTime) < cnt {
			cnt = len(uinsByAddFriendTime)
		}
		if len(uinsByPVCnt) < cnt {
			cnt = len(uinsByPVCnt)
		}

		combinOptions, err1 := GetOptionsByCombine(uin, uuid, randomUins, nextQGender, 4-cnt)
		if err1 != nil {
			err = err1
			log.Errorf(err.Error())
			return
		}

		if nextQGender == 1 {
			log.Debugf("uin %d, boyCnt(%d) < 4 GetOptionsByCombine %+v", uin, len(randomUins), combinOptions)
		} else if nextQGender == 2 {
			log.Debugf("uin %d, girlCnt(%d) < 4 GetOptionsByCombine %+v", uin, len(randomUins), combinOptions)
		} else {
			log.Debugf("uin %d, friendsCnt(%d) < 4  GetOptionsByCombine %+v", uin, len(randomUins), combinOptions)
		}

		//我的好友数据
		for _, uid := range randomUins {

			option := &st.OptionInfo2{uid, friendInfos[uid].NickName, "", uinsVoteCntMap[int(uid)]}
			options = append(options, option)
			prepared += fmt.Sprintf("%d:", uid)
		}

		uids := make([]int64, 0)
		//单向加好友或者通讯录或者明星
		for _, option := range combinOptions {
			options = append(options, option)

			if option.Uin != 0 {
				uids = append(uids, option.Uin)
				prepared += fmt.Sprintf("%d:", option.Uin)
			}
		}
		if len(prepared) > 0 {
			prepared = prepared[:len(prepared)-1]
		}

		log.Debugf("prepared:%+v", prepared)

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

		log.Debugf(" end  prepare candidates list options:%+v", options)
		return
	}

	//根据获取加好友的来源来进行判断分析
	//通讯录 -> 搜索 -> 同校同年级 -> 同校非同年级 -> 可能认识的人
	friendsByaddFriendSrc, err1 := GetUinsByAddFriendSrc(uin)
	if err1 != nil {
		log.Errorf(err1.Error())
	}

	uinsByVote = ReOrderUinsByAddFriendSrc(uin, uinsByVote, friendsByaddFriendSrc)
	uinsByVote, _ = OptionsByProbaility(uin, uinsByVote) // /将用户这一轮次答题中出现过的候选人往后排
	if len(uinsByVote) > 12 {
		uinsByVote = uinsByVote[:12]
	}

	uinsByAddFriendTime = ReOrderUinsByAddFriendSrc(uin, uinsByAddFriendTime, friendsByaddFriendSrc)
	uinsByAddFriendTime, _ = OptionsByProbaility(uin, uinsByAddFriendTime)
	if len(uinsByAddFriendTime) > 12 {
		uinsByAddFriendTime = uinsByAddFriendTime[:12]
	}

	uinsByPVCnt = ReOrderUinsByAddFriendSrc(uin, uinsByPVCnt, friendsByaddFriendSrc)
	uinsByPVCnt, _ = OptionsByProbaility(uin, uinsByPVCnt)
	if len(uinsByPVCnt) > 12 {
		uinsByPVCnt = uinsByPVCnt[:12]
	}

	randomUins = ReOrderUinsByAddFriendSrc(uin, randomUins, friendsByaddFriendSrc)
	randomUins, _ = OptionsByProbaility(uin, randomUins)
	if len(randomUins) > 12 {
		randomUins = randomUins[:12]
	}

	//log.Debugf("uin %d, friendUins %d, uinsByVote %+v,  uinsByAddFriendTime %+v, uinsByPVCnt %+v", uin, len(friendUins), uinsByVote, uinsByAddFriendTime, uinsByPVCnt)

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
		option := &st.OptionInfo2{uid, friendInfos[uid].NickName, "", uinsVoteCntMap[int(uid)]}
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

func OptionsByProbaility(uin int64, uids []int64) (newUids []int64, err error) {
	log.Errorf("start OptionsByProbaility uin:%d", uin)

	if uin == 0 || len(uids) == 0 {
		log.Errorf("uids is empty")
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_LOOKED_OPTION_UINS)
	if err != nil {
		log.Errorf(err.Error())

		//如果出错 保持原有顺序
		newUids = uids
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	vals, err := app.HGetAll(keyStr)
	if err != nil {
		log.Errorf(err.Error())

		//如果出错 保持原有顺序
		newUids = uids
		return
	}

	log.Errorf("uin = %d, vals:%+v", uin, vals)

	//按出现次数降序排序
	rvals := make(map[int]int)

	//所有的出现过的uin列表
	for k, v := range vals {
		ik, _ := strconv.ParseInt(k, 10, 64)
		iv, _ := strconv.ParseInt(v, 10, 32)
		rvals[int(ik)] = int(iv)
	}

	//按好友被看到最少的在最前
	tmpUids1 := make([]int64, 0)

	pairs := util.SortMap1(rvals)
	for _, p := range pairs {
		tmpUids1 = append(tmpUids1, int64(p.Key))
	}

	//找出在本次题目候选人集合和在用户这一轮答题候选人列表中都出现过的所有人
	tmpUids2 := make([]int64, 0)
	for _, uid := range tmpUids1 {
		find := false
		for _, u := range uids {
			if u == uid {
				find = true
				break
			}
		}

		if find {
			tmpUids2 = append(tmpUids2, uid)
		}
	}

	//将出现过的人排在本次题目候选人集合最后
	for _, uid := range uids {
		find := false
		for _, u := range tmpUids2 {
			if u == uid {
				find = true
				break
			}
		}

		if !find {
			newUids = append(newUids, uid)
		}
	}
	newUids = append(newUids, tmpUids2...)

	log.Errorf("end OptionsByProbaility uin:%d", uin)
	return
}

func ClearOptionUinsLookedCnt(uin int64) (err error) {
	log.Errorf("start ClearOptionUinsLookedCnt uin:%d", uin)
	if uin == 0 {
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_LOOKED_OPTION_UINS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	err = app.Del(keyStr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("end ClearOptionUinsLookedCnt uin:%d", uin)
	return
}

func AddOptionUinsLookedCnt(uin int64, options []*st.OptionInfo2) (err error) {

	log.Errorf("start AddOptionUinsLookedCnt uin:%d", uin)
	if len(options) == 0 || uin == 0 {
		return
	}

	optionUins := make([]int64, 0)

	for _, option := range options {
		if option.Uin > 0 {
			optionUins = append(optionUins, option.Uin)
		}
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_LOOKED_OPTION_UINS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	for _, uid := range optionUins {
		if uid == 0 {
			continue
		}

		fieldStr := fmt.Sprintf("%d", uid)

		_, err = app.HIncrBy(keyStr, fieldStr, 1)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

	}
	log.Errorf("end AddOptionUinsLookedCnt uin:%d", uin)
	return
}
