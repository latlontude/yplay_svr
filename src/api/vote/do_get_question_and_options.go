package vote

import (
	"common/constant"
	"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"common/token"
	"common/util"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"svr/cache"
	"svr/st"
	"time"
)

type GetQuestionAndOptionsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
}

type GetQuestionAndOptionsRsp struct {
	FreezeStatus   int               `json:"freezeStatus"`   //1 表示是冷冻状态, 0 表示非冷冻状态
	FreezeTs       int               `json:"freezeTs"`       //冷冻开始的时间，对于冷冻状态有效，非冷冻状态时为0
	NowTs          int               `json:"nowTs"`          //服务器当前时间
	FreezeDuration int               `json:"freezeDuration"` //冷冻时长 服务器配置
	Total          int               `json:"total"`          //问题总数
	Index          int               `json:"index"`          //问题编号
	Question       *st.QuestionInfo  `json:"question"`       //当前题目列表
	Options        []*st.OptionInfo2 `json:"options"`        //选项
}

func doGetQuestionAndOptions(req *GetQuestionAndOptionsReq, r *http.Request) (rsp *GetQuestionAndOptionsRsp, err error) {

	log.Debugf("uin %d, GetQuestionAndOptionsReq %+v", req.Uin, req)

	uuid, err := token.GetUuidFromTokenString(req.Token, req.Ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt "+err.Error())
		log.Errorf("uin %d, GetQuestionAndOptionsRsp error, %s", req.Uin, err.Error())
		return
	}

	//默认为非冷冻状态
	freezeStatus := constant.ENUM_FREEZE_STATUS_NOT_FROZEND

	//查询上次冷冻状态时间
	freezeTs, err := st.GetFreezingStatus(req.Uin)
	if err != nil {
		log.Errorf("uin %d, GetQuestionAndOptionsRsp error, %s", req.Uin, err.Error())
		return
	}

	freezeDuration := env.Config.Vote.FreezeDuration

	//处于冷冻状态 不返还任何问题
	nowTs := int(time.Now().Unix())
	if nowTs-freezeTs < freezeDuration {
		freezeStatus = constant.ENUM_FREEZE_STATUS_FROZEND

		rsp = &GetQuestionAndOptionsRsp{freezeStatus, freezeTs, nowTs, freezeDuration, constant.ENUM_QUESTION_BATCH_SIZE, 0, nil, make([]*st.OptionInfo2, 0)}

		log.Debugf("uin %d, GetQuestionAndOptionsRsp succ, %+v", req.Uin, rsp)
		return
	}

	var qinfo *st.QuestionInfo
	var options []*st.OptionInfo2
	var index int

	//if req.Uin == 100328 || req.Uin == 100446 {
	if true {
		qinfo, options, index, err = GetNextQuestionAndOptionsByPreGene(req.Uin, uuid)
	} else {
		qinfo, options, index, err = GetNextQuestionAndOptions(req.Uin, uuid)
	}

	if err != nil {
		log.Errorf("uin %d, GetQuestionAndOptionsRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &GetQuestionAndOptionsRsp{freezeStatus, freezeTs, nowTs, freezeDuration, constant.ENUM_QUESTION_BATCH_SIZE, index, qinfo, options}

	log.Debugf("uin %d, GetQuestionAndOptionsRsp succ, %+v", req.Uin, rsp)

	return
}

func GetNextQuestionAndOptions(uin int64, uuid int64) (qinfo *st.QuestionInfo, options []*st.OptionInfo2, index int, err error) {

	log.Debugf("GetNextQuestionAndOptions begin, uin %d, uuid %d", uin, uuid)

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_QINFO)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"qid", "qindex", "options", "voted", "cursor1", "cursor2", "cursor3"}

	optionsStr := ""

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
	lastQGender := 2 //上一次的题目性别
	lastQIndex := 0  //上一次的答题编号 1~15
	voted := 1       //是否已经投票
	lastCursor1 := 0 //三个队列的上次扫描位置
	lastCursor2 := 0 //三个队列的上次扫描位置
	lastCursor3 := 0 //三个队列的上次扫描位置

	//如果从来没有答题 则上一次题目设置为0 上一次的性别设置为2(下一次从通用开始)，上答题一次索引为0
	if len(valsStr) != 0 {
		lastQId, _ = strconv.Atoi(valsStr["qid"])
		lastQIndex, _ = strconv.Atoi(valsStr["qindex"])
		voted, _ = strconv.Atoi(valsStr["voted"])
		optionsStr = valsStr["options"]

		lastCursor1, _ = strconv.Atoi(valsStr["cursor1"])
		lastCursor2, _ = strconv.Atoi(valsStr["cursor2"])
		lastCursor3, _ = strconv.Atoi(valsStr["cursor3"])
	}

	log.Debugf("uin %d, lastQId %d, lastQIndex %d, voted %d, cursor1 %d, cursor2 %d, cursor3 %d, optionsStr %s", uin, lastQId, lastQIndex, voted, lastCursor1, lastCursor2, lastCursor3, optionsStr)

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

	index = lastQIndex + 1
	if index > 15 {
		index = 1
	}

	//如果在数据库中修改了题目原有的性别,以新的性别为准
	if qi, ok := cache.QUESTIONS[lastQId]; ok {
		lastQGender = qi.OptionGender
	}

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

	nextQGender := 0

	//下一道题的性别
	if boyCnt >= constant.ENUM_OPTION_BATCH_SIZE && girlCnt >= constant.ENUM_OPTION_BATCH_SIZE {

		//0->1
		//1->2
		//2->0
		nextQGender = (lastQGender + 1) % 3

	} else if boyCnt >= constant.ENUM_OPTION_BATCH_SIZE {

		//0->1
		//1->0
		//2->0
		if lastQGender > 0 {
			nextQGender = 0
		} else {
			nextQGender = 1
		}

	} else if girlCnt >= constant.ENUM_OPTION_BATCH_SIZE {

		//0->2
		//1->2
		//2->0
		if lastQGender == 2 {
			nextQGender = 0
		} else {
			nextQGender = 2
		}

	} else {
		// boyCnt < 4 && girlCnt < 4
		nextQGender = 0
	}

	lastCursor := lastCursor1
	if nextQGender == 1 {
		lastCursor = lastCursor2
	}

	if nextQGender == 2 {
		lastCursor = lastCursor3
	}

	log.Debugf("uin %d, boyCnt %d, girlCnt %d, lastQGender %d, nextQGender %d, lastCursor %d", uin, boyCnt, girlCnt, lastQGender, nextQGender, lastCursor)

	//找到下一题
	newQId, err := GetNextQIdByGender(uin, nextQGender, lastCursor)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, genen newQId %d", uin, newQId)

	qinfo = cache.QUESTIONS[newQId]

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

			log.Debugf("uin %d, GetOptions, in defer func, options %+v", uin, options)

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

			if nextQGender == 0 {
				lastCursor1 = newQId
			} else if nextQGender == 1 {
				lastCursor2 = newQId
			} else {
				lastCursor3 = newQId
			}

			kvs["cursor1"] = fmt.Sprintf("%d", lastCursor1)
			kvs["cursor2"] = fmt.Sprintf("%d", lastCursor2)
			kvs["cursor3"] = fmt.Sprintf("%d", lastCursor3)

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

	//根据获取加好友的来源来进行判断分析
	//通讯录 -> 搜索 -> 同校同年级 -> 同校非同年级 -> 可能认识的人
	friendsByaddFriendSrc, err1 := GetUinsByAddFriendSrc(uin)
	if err1 != nil {
		log.Errorf(err1.Error())
	}

	if len(uinsByVote) > 12 {
		uinsByVote = ReOrderUinsByAddFriendSrc(uinsByVote, friendsByaddFriendSrc)
		uinsByVote = uinsByVote[:12]
	}

	if len(uinsByAddFriendTime) > 12 {
		uinsByAddFriendTime = ReOrderUinsByAddFriendSrc(uinsByAddFriendTime, friendsByaddFriendSrc)
		uinsByAddFriendTime = uinsByAddFriendTime[:12]
	}

	if len(uinsByPVCnt) > 12 {
		uinsByPVCnt = ReOrderUinsByAddFriendSrc(uinsByPVCnt, friendsByaddFriendSrc)
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

func PrepareAllOptionsUin2(uin int64, uinsByVote, uinsByAddFriendTime, uinsByPVCnt, randomUins []int64) (allOptionUins []int64, err error) {

	if len(uinsByVote) < 4 || len(uinsByAddFriendTime) < 4 || len(uinsByPVCnt) < 4 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "GetAllOptionsUin invalid param")
		log.Errorf(err.Error())
		return
	}

	allOptionUins = make([]int64, 0)

	selUins1 := make([]int64, 0) //答题最多  4个
	selUins2 := make([]int64, 0) //加好友时间+钻石数 5个
	selUin3 := int64(0)          //按用户活跃数最少  1个
	selUins4 := make([]int64, 0) //随机用户 2个

	orders1 := rand.Perm(len(uinsByVote))          //随机化
	orders2 := rand.Perm(len(uinsByAddFriendTime)) //随机化

	//找4个
	selUins1 = append(selUins1, uinsByVote[orders1[0]])
	selUins1 = append(selUins1, uinsByVote[orders1[1]])
	selUins1 = append(selUins1, uinsByVote[orders1[2]])
	selUins1 = append(selUins1, uinsByVote[orders1[3]])

	log.Debugf("uin %d, PrepareAllOptionsUin2 selUins1 %+v", uin, selUins1)

	//找5个
	i := 0
	for {

		if i >= len(uinsByAddFriendTime) {
			break
		}

		t := uinsByAddFriendTime[orders2[i]]

		find := false

		for _, uid := range selUins1 {
			if t == uid {
				find = true
				break
			}
		}

		if !find {
			selUins2 = append(selUins2, t)

			if len(selUins2) >= 5 {
				break
			}
		}

		i++

	}

	log.Debugf("uin %d, PrepareAllOptionsUin2 selUins2 %+v", uin, selUins2)

	//找1个
	for _, t := range uinsByPVCnt {

		find := false

		for _, uid := range selUins1 {
			if t == uid {
				find = true
				break
			}
		}

		if find {
			continue
		}

		for _, uid := range selUins2 {
			if t == uid {
				find = true
				break
			}
		}

		if !find {
			selUin3 = t
			break
		}
	}

	log.Debugf("uin %d, PrepareAllOptionsUin2 selUin3 %+v", uin, selUin3)

	curUins := make([]int64, 0)

	for _, uid := range selUins1 {
		curUins = append(curUins, uid)
	}

	for _, uid := range selUins2 {
		curUins = append(curUins, uid)
	}

	if selUin3 > 0 {
		curUins = append(curUins, selUin3)
	}

	//找2个随机
	orders3 := rand.Perm(len(randomUins))

	for _, order := range orders3 {

		uid := randomUins[order]

		find := false

		for _, t := range curUins {
			if t == uid {
				find = true
				break
			}
		}

		if find {
			continue
		} else {
			selUins4 = append(selUins4, uid)
			if len(selUins4) >= 2 {
				break
			}
		}
	}

	log.Debugf("uin %d, PrepareAllOptionsUin2 selUins4 %+v", uin, selUins4)

	//第一批2个最合适 + 1个钻石少 + 1个随机好友
	allOptionUins = append(allOptionUins, selUins1[0])
	allOptionUins = append(allOptionUins, selUins1[1])
	if len(selUins2) >= 1 {
		allOptionUins = append(allOptionUins, selUins2[0])
	}
	if len(selUins4) >= 1 {
		allOptionUins = append(allOptionUins, selUins4[0])
	}

	//第二批1个最合适 + 2个钻石少 + 1个随机好友
	allOptionUins = append(allOptionUins, selUins1[2])
	if len(selUins2) >= 2 {
		allOptionUins = append(allOptionUins, selUins2[1])
	}
	if len(selUins2) >= 3 {
		allOptionUins = append(allOptionUins, selUins2[2])
	}
	if len(selUins4) >= 2 {
		allOptionUins = append(allOptionUins, selUins4[1])
	}

	//第三批1个最合适 + 2个钻石少 + 1个活跃值
	allOptionUins = append(allOptionUins, selUins1[3])
	if len(selUins2) >= 4 {
		allOptionUins = append(allOptionUins, selUins2[3])
	}
	if len(selUins2) >= 5 {
		allOptionUins = append(allOptionUins, selUins2[4])
	}
	if selUin3 > 0 {
		allOptionUins = append(allOptionUins, selUin3)
	}

	return
}

//要求下一题的性别及上一次答题的游标
//三个队列
//ALL_GENE_QIDS 通用题目
//ALL_BOY_QIDS  男性题目
//ALL_GIRL_QIDS 女性题目
func GetNextQIdByGender(uin int64, qgender int, cursor int) (qid int, err error) {

	log.Debugf("uin %d, GetNextQIdByGender gender %d, cursor %d", uin, qgender, cursor)

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	ui, err := st.GetUserProfileInfo(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	var qids []int

	//无性别要求
	if qgender == 0 {
		qids = cache.ALL_GENE_QIDS
	}

	//要求是男性题目
	if qgender == 1 {
		qids = cache.ALL_BOY_QIDS
	}

	//要求是女性题目
	if qgender == 2 {
		qids = cache.ALL_GIRL_QIDS
	}

	//如果上次的游标是最后一个，则重置
	if cursor >= qids[len(qids)-1] {
		cursor = 0
	}

	find := false
	for _, q := range qids {
		if q > cursor {

			qi := cache.QUESTIONS[q]

			//答题者性别要求或者无性别要求
			if (qi.ReplyGender == ui.Gender || qi.ReplyGender == 0) && (qi.SchoolType == 0 || ui.SchoolType == 0 || (qi.SchoolType&ui.SchoolType > 0)) {
				//if qi.ReplyGender == ui.Gender || qi.ReplyGender == 0 {
				find = true
				qid = q
				break
			}
		}
	}

	log.Debugf("uin %d, GetNextQIdByGender already scan all, begin new cycle", uin)

	//如果循环已经到底
	if !find {

		for _, q := range qids {

			qi := cache.QUESTIONS[q]

			//答题者性别要求或者无性别要求
			if qi.ReplyGender == ui.Gender || qi.ReplyGender == 0 {
				find = true
				qid = q
				break
			}
		}
	}

	if !find {
		err = rest.NewAPIError(constant.E_RES_NOT_FOUND, "can't find next qid by logic")
		log.Errorf(err.Error())
		return
	}

	return
}

//从单向加好友->通讯录好友->固定一批名人
//qgender必定等于0
//没有计算这些人被选中的次数
func GetOptionsByCombine(uin int64, uuid int64, excludeUins []int64, qgender int, cnt int) (options []*st.OptionInfo2, err error) {

	if uin == 0 || cnt == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err.Error())
		return
	}

	//从单向添加好友中获取
	options, err = GetOptionsFromAddFriendMsg(uin, cnt)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//检查是否有重复的
	noptions := make([]*st.OptionInfo2, 0)
	for _, option := range options {

		find := false
		for _, uid := range excludeUins {

			if option.Uin == uid {
				find = true
				break
			}
		}

		if !find {
			noptions = append(noptions, option)
		}
	}

	options = noptions

	newExcludeUins := excludeUins

	for _, option := range options {
		if option.Uin != 0 {
			newExcludeUins = append(newExcludeUins, option.Uin)
		}
	}

	if len(options) >= cnt {
		return
	}

	needCnt := cnt - len(options)

	log.Debugf("uin %d, GetOptionsFromAddrBook needCnt %d, uuid %d, excludeUins %+v", uin, needCnt, uuid, newExcludeUins)

	//从通讯录中获取
	options2, err := GetOptionsFromAddrBook(uin, uuid, newExcludeUins, needCnt)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, GetOptionsFromAddrBook cnt %d, rsp %+v", uin, len(options2), options2)

	for _, opt := range options2 {
		options = append(options, opt)
	}

	if len(options) >= cnt {
		return
	}

	//从指定的明星中获取
	needCnt = cnt - len(options)

	base := 4
	if qgender == 0 {
		base = 8
	}

	midxs := make(map[int]int)
	for {

		idx := rand.Intn(base)
		midxs[idx] = 1

		if len(midxs) >= needCnt {
			break
		}
	}

	ALL_STARS := []string{"吴亦凡", "王思聪", "刘昊然", "薛之谦", "李宇春", "周冬雨", "迪丽热巴", "杨幂"}
	ALL_GIRL_STARS := []string{"李宇春", "周冬雨", "迪丽热巴", "杨幂"}
	ALL_BOY_STARS := []string{"吴亦凡", "王思聪", "刘昊然", "薛之谦"}

	//qgender必定为0
	for idx, _ := range midxs {

		nickName := ""

		if qgender == 0 {
			nickName = ALL_STARS[idx]
		}

		if qgender == 1 {
			nickName = ALL_BOY_STARS[idx]
		}

		if qgender == 2 {
			nickName = ALL_GIRL_STARS[idx]
		}

		option := &st.OptionInfo2{0, nickName, 0}

		options = append(options, option)
	}

	return
}

func GetOptionsFromAddFriendMsg(uin int64, cnt int) (options []*st.OptionInfo2, err error) {

	if uin == 0 || cnt == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	log.Debugf("uin %d, GetOptionsFromAddFriendMsg need cnt %d", uin, cnt)

	//获取用户在当前题目下被选中的次数
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	options = make([]*st.OptionInfo2, 0)

	//检查是否存在这样的消息 必须不是好友
	sql := fmt.Sprintf(`select count(toUin) from addFriendMsg where fromUin = %d and status != 1`, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		rows.Scan(&total)
	}

	log.Debugf("uin %d, GetOptionsFromAddFriendMsg, total %d", uin, total)

	if total == 0 {
		return
	}

	s := 0

	//从单向添加过的好友里面满足所有的信息
	if total >= cnt {
		s = rand.Intn(total - cnt + 1)
	} else {
		cnt = total
	}

	sql = fmt.Sprintf(`select toUin from addFriendMsg where fromUin = %d and status != 1 limit %d,%d`, uin, s, cnt)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}

	uins := make([]int64, 0)
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)

		uins = append(uins, uid)
	}

	res, err := st.BatchGetUserProfileInfo(uins)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, v := range res {

		if len(v.NickName) == 0 {
			continue
		}

		option := &st.OptionInfo2{v.Uin, v.NickName, 0}
		options = append(options, option)
	}

	log.Debugf("uin %d, GetOptionsFromAddFriendMsg, options %+v", uin, options)

	return
}

//从通讯录随机选取，尽可能的返回，可能不足
func GetOptionsFromAddrBook(uin, uuid int64, excludeUins []int64, needCnt int) (options []*st.OptionInfo2, err error) {

	options = make([]*st.OptionInfo2, 0)

	if needCnt <= 0 {
		return
	}

	ops, err := GetOptionsFromAddrBookRegister(uin, uuid, excludeUins, needCnt)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, op := range ops {
		options = append(options, op)
	}

	if len(options) >= needCnt {
		return
	}

	ops2, err := GetOptionsFromAddrBookUnRegister(uin, uuid, needCnt-len(options))
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, op := range ops2 {
		options = append(options, op)
	}

	return
}

//从通讯录随机选取，尽可能的返回，可能不足
func GetOptionsFromAddrBookRegister(uin, uuid int64, excludeUins []int64, needCnt int) (options []*st.OptionInfo2, err error) {

	options = make([]*st.OptionInfo2, 0)

	if needCnt <= 0 {
		return
	}

	//获取用户在当前题目下被选中的次数
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	strs := ""
	for _, uid := range excludeUins {
		strs += fmt.Sprintf("%d,", uid)
	}
	strs += fmt.Sprintf("%d,", uin)
	strs += fmt.Sprintf("%d", 0) //查找注册好友

	sql := fmt.Sprintf(`select count(friendUin) from addrBook where uuid = %d and friendUin not in (%s)`, uuid, strs)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var total int
	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	//实际总数比需要的要少
	if needCnt >= total {
		needCnt = total
	}

	idxs := GetRandomIdxs(total, needCnt)

	validUins := make([]int64, 0)

	for _, idx := range idxs {

		sql = fmt.Sprintf(`select friendUin, friendName from addrBook where uuid = %d and friendUin not in (%s) limit %d, %d`, uuid, strs, idx, 1)

		rows, err = inst.Query(sql)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
			log.Error(err)
			return
		}

		defer rows.Close()

		for rows.Next() {
			var option st.OptionInfo2
			rows.Scan(&option.Uin, &option.NickName)

			options = append(options, &option)

			if option.Uin > 0 {
				validUins = append(validUins, option.Uin)
			}
		}
	}

	//从通讯录里面选取的昵称要改成注册用户的正式昵称
	//默认通讯录好友有昵称
	res, _ := st.BatchGetUserProfileInfo(validUins)
	for i, option := range options {

		if ui, ok := res[option.Uin]; ok {

			//如果个人资料的昵称为空，则不替换
			if len(ui.NickName) > 0 {
				options[i].NickName = ui.NickName
			}
		}
	}

	return
}

//从通讯录随机选取，尽可能的返回，可能不足
func GetOptionsFromAddrBookUnRegister(uin, uuid int64, needCnt int) (options []*st.OptionInfo2, err error) {

	options = make([]*st.OptionInfo2, 0)

	if needCnt <= 0 {
		return
	}

	//获取用户在当前题目下被选中的次数
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select count(friendUin) from addrBook where uuid = %d and friendUin = 0`, uuid)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	var total int
	for rows.Next() {
		rows.Scan(&total)
	}

	if total == 0 {
		return
	}

	//实际总数比需要的要少
	if needCnt >= total {
		needCnt = total
	}

	idxs := GetRandomIdxs(total, needCnt)

	for _, idx := range idxs {

		sql = fmt.Sprintf(`select friendUin, friendName from addrBook where uuid = %d and friendUin = 0 limit %d, %d`, uuid, idx, 1)

		rows, err = inst.Query(sql)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
			log.Error(err)
			return
		}

		defer rows.Close()

		for rows.Next() {
			var option st.OptionInfo2
			rows.Scan(&option.Uin, &option.NickName)

			options = append(options, &option)
		}
	}

	return
}

func GetRandomIdxs(total int, cnt int) (idxs []int) {

	idxs = make([]int, 0)

	if total < cnt || total <= 0 || cnt <= 0 {
		return
	}

	if total <= 20 {
		a := rand.Perm(total)

		for _, idx := range a {
			idxs = append(idxs, idx)

			if len(idxs) == cnt {
				break
			}
		}

		return
	}

	idxM := make(map[int]int)

	//cnt比较小的情况下 比较容易退出。否则cnt很大 并且比较接近total，可能循环多次退不出
	for {

		n := rand.Intn(total)
		idxM[n] = 1

		if len(idxM) == cnt {
			break
		}
	}

	for idx, _ := range idxM {
		idxs = append(idxs, idx)
	}

	return
}

func GetUinsGemCnt(uins []int64) (res map[int64]int, err error) {

	res = make(map[int64]int)

	if len(uins) == 0 {
		return
	}

	str := ``
	for i, uin := range uins {

		if i != len(uins)-1 {
			str += fmt.Sprintf(`%d,`, uin)
		} else {
			str += fmt.Sprintf(`%d`, uin)
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select uin, statValue from userStat where uin in (%s) and statField = %d`, str, constant.ENUM_USER_STAT_GEM_CNT)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid int64
		var val int
		rows.Scan(&uid, &val)
		res[uid] = val
	}

	return
}

func GetUinsPVCnt(uins []int64) (res map[int]int, err error) {
	log.Errorf("start GetUinsPVCnt")
	res = make(map[int]int)

	if len(uins) == 0 {
		return
	}

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_USER_PV_CNT)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keys := make([]string, 0)

	y1, w1 := time.Now().ISOWeek()
	y2, w2 := time.Now().Add(-604800 * time.Second).ISOWeek()  //1周前
	y3, w3 := time.Now().Add(-1209600 * time.Second).ISOWeek() //2周前

	for _, uin := range uins {
		keys = append(keys, fmt.Sprintf("%d_%d_%d", uin, y1, w1))
		keys = append(keys, fmt.Sprintf("%d_%d_%d", uin, y2, w2))
		keys = append(keys, fmt.Sprintf("%d_%d_%d", uin, y3, w3))
	}

	r, err := app.MGet(keys)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("ret : %+v", r)

	for _, uin := range uins {

		res[int(uin)] = 0

		key1 := fmt.Sprintf("%d_%d_%d", uin, y1, w1)
		key2 := fmt.Sprintf("%d_%d_%d", uin, y2, w2)
		key3 := fmt.Sprintf("%d_%d_%d", uin, y3, w3)

		if v, ok := r[key1]; ok {
			t, _ := strconv.Atoi(v)
			res[int(uin)] += t
		}

		if v, ok := r[key2]; ok {
			t, _ := strconv.Atoi(v)
			res[int(uin)] += t
		}

		if v, ok := r[key3]; ok {
			t, _ := strconv.Atoi(v)
			res[int(uin)] += t
		}
	}
	log.Errorf("end GetUinsPVCnt res:%+v", res)
	return
}

func GetUinsByAddFriendSrc(uin int64) (uinsMap map[int64]int, err error) {
	log.Errorf("start GetUinsByAddFriendSrc uin:%d", uin)
	uinsMap = make(map[int64]int)

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	sql := fmt.Sprintf(`select fromUin, toUin, srcType, mts from addFriendMsg where fromUin = %d or toUin = %d`, uin, uin)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		return
	}
	defer rows.Close()

	uids1 := make([]int64, 0)
	uids2 := make([]int64, 0)
	uids3 := make([]int64, 0)
	uids4 := make([]int64, 0)
	uids5 := make([]int64, 0)

	friendUinsMap := make(map[int64]int)
	for rows.Next() {
		var fromUin int64
		var toUin int64
		var srcType int
		var mts int
		var uid int64
		rows.Scan(&fromUin, &toUin, &srcType, &mts)

		if mts == 0 { // 发送了添加好友请求，但目前还不是好友
			continue
		}
		if fromUin == uin {
			uid = toUin
		} else {
			uid = fromUin
		}

		friendUinsMap[uid] = srcType
	}

	for uid, srcType := range friendUinsMap {
		if srcType == 1 {
			//通讯录好友
			uids1 = append(uids1, uid)
		} else if srcType == 8 {
			//搜索好友
			uids2 = append(uids2, uid)
		} else if srcType == 4 {
			//同校同年级
			uids3 = append(uids3, uid)
		} else if srcType == 3 || srcType == 5 || srcType == 6 {
			//同校/同校男生/同校女生/
			uids4 = append(uids4, uid)
		} else if srcType == 7 {
			//共同好友
			uids5 = append(uids5, uid)
		}
	}

	log.Errorf("uids1:%+v, uids2:%+v, uids3:%+v, uids4:%+v, uids5:%+v", uids1, uids2, uids3, uids4, uids5)

	for _, uid := range uids1 {
		uinsMap[uid] = 1
	}
	for _, uid := range uids2 {
		uinsMap[uid] = 2
	}
	for _, uid := range uids3 {
		uinsMap[uid] = 3
	}
	for _, uid := range uids4 {
		uinsMap[uid] = 4
	}
	for _, uid := range uids5 {
		uinsMap[uid] = 5
	}

	log.Errorf("end GetUinsByAddFriendSrc uinsMap:%+v", uinsMap)
	return
}

func ReOrderUinsByAddFriendSrc(uins []int64, uinsByaddFriendSrc map[int64]int) (ordered []int64) {

	log.Errorf("start ReOrderUinsByAddFriendSrc  uins:%+v", uins)
	//按照friendsByAddFriendSrc的顺序
	ordered = make([]int64, 0)

	uids1 := make([]int64, 0)
	uids2 := make([]int64, 0)
	uids3 := make([]int64, 0)
	uids4 := make([]int64, 0)
	uids5 := make([]int64, 0)
	uids6 := make([]int64, 0) // 不能判断添加好友来源的好友列表

	for _, uid := range uins {
		if src, ok := uinsByaddFriendSrc[uid]; ok {
			switch src {
			case 1:
				uids1 = append(uids1, uid)
			case 2:
				uids2 = append(uids2, uid)
			case 3:
				uids3 = append(uids3, uid)
			case 4:
				uids4 = append(uids4, uid)
			case 5:
				uids5 = append(uids5, uid)
			default:
				log.Errorf("wrong src : %d", src)
			}
		} else {
			uids6 = append(uids6, uid)
		}
	}

	ordered = append(ordered, uids1...)
	ordered = append(ordered, uids2...)
	ordered = append(ordered, uids3...)
	ordered = append(ordered, uids4...)
	ordered = append(ordered, uids5...)
	ordered = append(ordered, uids6...)

	log.Errorf("end ReOrderUinsByAddFriendSrc ordered:%+v", ordered)
	return
}
