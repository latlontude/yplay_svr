package vote

import (
	"common/constant"
	"common/myredis"
	"common/rest"
	"common/token"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"svr/st"
)

type GetOptionsReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`
	QId   int    `schema:"qid"`
	Index int    `schema:"index"`
}

type GetOptionsRsp struct {
	Options []*st.OptionInfo2 `json:"options"` //选项
}

func doGetOptions(req *GetOptionsReq, r *http.Request) (rsp *GetOptionsRsp, err error) {

	log.Debugf("uin %d, qid %d, GetOptionsReq %+v", req.Uin, req.QId, req)

	uuid, err := token.GetUuidFromTokenString(req.Token, req.Ver)
	if err != nil {
		err = rest.NewAPIError(constant.E_INVALID_SESSION, "token decrypt "+err.Error())
		log.Errorf("uin %d, qid %d, GetOptionsRsp error, %s", req.Uin, req.QId, err.Error())
		return
	}

	var options []*st.OptionInfo2
	//if req.Uin == 100328 || req.Uin == 100446 {
	if true {
		options, err = GetOptionsByPreGene(req.Uin, req.QId, req.Index, uuid)
	} else {
		options, err = GetOptions(req.Uin, req.QId, req.Index, uuid)
	}

	if err != nil {
		log.Errorf("uin %d, qid %d, GetOptionsRsp error, %s", req.Uin, req.QId, err.Error())
		return
	}

	rsp = &GetOptionsRsp{options}

	log.Debugf("uin %d, qid %d, GetOptionsRsp succ, %+v", req.Uin, req.QId, rsp)

	return
}

func GetOptions(uin int64, qid int, index int, uuid int64) (options []*st.OptionInfo2, err error) {

	options = make([]*st.OptionInfo2, 0)

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_LAST_QINFO)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d", uin)

	//上一次答题的ID 上一次题目的性别 上一次答题的索引
	fields := []string{"qid", "qindex", "voted", "prepared", "preparedcursor"}

	valsStr, err := app.HMGet(keyStr, fields)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Debugf("GetOptions uin %d, HMGet rsp %+v", uin, valsStr)

	if len(valsStr) != len(fields) || len(valsStr) == 0 {
		err = rest.NewAPIError(constant.E_REDIS_HASH, "redis ret hash size")
		log.Errorf(err.Error())
		return
	}

	lastQId, _ := strconv.Atoi(valsStr["qid"])
	lastQIndex, _ := strconv.Atoi(valsStr["qindex"])
	voted, _ := strconv.Atoi(valsStr["voted"])
	prepared := valsStr["prepared"]
	preparedCursor, _ := strconv.Atoi(valsStr["preparedcursor"])

	log.Debugf("GetOptions uin %d, lastQId %d, lastQIndex %d, voted %d, prepared %s, preparedCursor %d", uin, lastQId, lastQIndex, voted, prepared, preparedCursor)

	//上次题目已经回答
	if voted == 1 {
		//不可能出现的情况
		log.Errorf("uin %d, qid %d, index %d, getoption, invalid voted status", uin, qid, index)
	}

	if qid != lastQId || index != lastQIndex {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "qid or qindex not match")
		log.Errorf(err.Error())
		return
	}

	uids := make([]int64, 0)
	strUins := strings.Split(prepared, ":")
	for _, strUin := range strUins {
		uid, _ := strconv.Atoi(strUin)

		if uid == 0 {
			continue
		}

		uids = append(uids, int64(uid))
	}

	friendInfos, err := st.BatchGetUserProfileInfo(uids)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	uinsVoteCntMap, err := st.GetUinsVotedCnt(qid, uids)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

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
			optionsStr := string(optionsTStr)

			kvs["options"] = optionsStr
			kvs["preparedcursor"] = fmt.Sprintf("%d", preparedCursor)

			err = app.HMSet(keyStr, kvs)
			if err != nil {
				log.Error(err.Error())
				return
			}
		}
	}()

	//好友人数小于4人, 当前好友 + 单项添加过的好友 + 通讯录好友 + 默认补充
	if len(uids) < constant.ENUM_OPTION_BATCH_SIZE {

		combinOptions, err1 := GetOptionsByCombine(uin, uuid, uids, 0, 4-len(uids))
		if err1 != nil {
			err = err1
			log.Errorf(err.Error())
			return
		}

		log.Debugf("uin %d, friendUins(%d)<4,  GetOptionsByCombine %+v", uin, len(uids), combinOptions)

		//我的好友数据
		for _, uid := range uids {

			option := &st.OptionInfo2{uid, friendInfos[uid].NickName, uinsVoteCntMap[int(uid)]}
			options = append(options, option)
		}

		uidTmps := make([]int64, 0)
		//单向加好友或者通讯录或者明星
		for _, option := range combinOptions {
			options = append(options, option)

			if option.Uin != 0 {
				uidTmps = append(uidTmps, option.Uin)
			}
		}

		if len(uidTmps) == 0 {
			return
		}

		//单向加好友或者通讯录好友在该题目下被选择的次数
		voteCntMap, _ := st.GetUinsVotedCnt(qid, uidTmps)

		for i, option := range options {

			if v, ok := voteCntMap[int(option.Uin)]; ok {
				options[i].BeSelCnt = v
			}
		}

		return
	}

	selectedUins := make([]int64, 0)
	for {

		preparedCursor += 1
		selectedUins = append(selectedUins, uids[(preparedCursor)%len(uids)])

		if len(selectedUins) == 4 {
			break
		}
	}

	log.Debugf("uin %d, GetOptions selectedUins %+v", uin, selectedUins)

	for _, uid := range selectedUins {

		nickName := ""
		if fi, ok := friendInfos[uid]; ok {
			nickName = fi.NickName
		}

		option := &st.OptionInfo2{uid, nickName, uinsVoteCntMap[int(uid)]}
		options = append(options, option)
	}

	return
}
