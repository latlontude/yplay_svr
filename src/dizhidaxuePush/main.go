package main

import (
	"api/im"
	"bytes"
	"common/constant"
	"common/env"
	//"common/mydb"
	"common/myredis"
	"common/rest"
	//"common/util"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//计算2度好友关系的SVR配置
type DizhidaxuePushConfig struct {
	Log struct {
		LogPath     string
		LogFileName string
		LogLevel    string //"fatal,error,warning,info,debug"
	}

	//DbInsts    map[string]DataBase
	RedisInsts map[string]env.RedisInst
	RedisApps  map[string]env.RedisApp
}

var (
	confFile string
	config   DizhidaxuePushConfig

	uidsStr string
	log     = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/dizhidaxuepush.conf", "默认配置文件路径")
	flag.StringVar(&uidsStr, "u", "", "用户列表")
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func main() {

	flag.Parse()

	if len(uidsStr) == 0 {
		fmt.Printf("uids length 0!")
		return
	}

	a := strings.Split(uidsStr, ",")
	if len(a) == 0 {
		fmt.Printf("invalid uids[%s] string!", uidsStr)
		return
	}

	uids := make([]int, 0)
	for _, uidStr := range a {
		uid, _ := strconv.Atoi(uidStr)

		if uid == 0 {
			continue
		}

		uids = append(uids, uid)
	}

	fmt.Printf("uids %+v", uids)

	if len(uids) == 0 {
		fmt.Printf("invalid uids[%s] string!", uidsStr)
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &config))
	//panicUnless(mydb.Init(config.DbInsts))
	panicUnless(env.InitLog(config.Log.LogPath, config.Log.LogFileName, config.Log.LogLevel))
	panicUnless(myredis.Init(config.RedisInsts, config.RedisApps))

	log.Errorf("start.....")

	for _, uid := range uids {

		err := SendPushMsg(int64(uid))
		if err != nil {
			log.Errorf("%d, sendpushmsg error, [%s]", uid, err.Error())
			continue
		}

		/*
			err = InsertNewQid(uid)
			if err != nil {
				log.Errorf("%d, InsertNewQid error, [%s]", uid, err.Error())
				continue
			}

			log.Errorf("%d, allaction succ", uid)
		*/
	}

	time.Sleep(3 * time.Second)

}

func InsertNewQid(uin int) (err error) {

	//从redis获取上一次答题的信息
	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_PRE_GENE_QIDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	keyStr := fmt.Sprintf("%d_progress", uin)
	fields := []string{"cursor"}

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

	lastCursor := -1 //问题队列的上次扫描位置

	//如果从来没有答题 则上一次题目设置为0 上答题一次索引为0
	if len(valsStr) != 0 {
		lastCursor, _ = strconv.Atoi(valsStr["cursor"])
	}

	keyStr2 := fmt.Sprintf("%d_qids", uin)

	qid := 21151

	vals, err := app.ZRangeWithScores(keyStr2, lastCursor+1, lastCursor+1)
	if len(vals) != 2 || err != nil {
		return
	}

	//先删除原来的
	app.ZRemRangeByRank(keyStr2, lastCursor+1, lastCursor+1)

	//score
	orgScore, _ := strconv.Atoi(vals[1])

	log.Debugf("uin %d, lastCursor+1 %d, lastScore %+v", uin, lastCursor+1, orgScore)

	err = app.ZAdd(keyStr2, int64(orgScore), fmt.Sprintf("%d", qid))
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	return
}

func MakeIMPushMsg(uin int64) (msg im.IMC2CMsg, err error) {

	log.Debugf("begin MakeIMPushMsg uin %d", uin)

	var customData im.IMCustomData
	customData.DataType = 4
	customData.Data = ""

	var customContent im.IMCustomContent
	cc, _ := json.Marshal(customData)
	customContent.Data = string(cc)
	customContent.Desc = ""
	customContent.Ext = ""
	customContent.Sound = ""

	var leaveFrozenMsgBody im.IMMsgBody
	leaveFrozenMsgBody.MsgType = "TIMCustomElem"
	leaveFrozenMsgBody.MsgContent = customContent

	msg.SyncOtherMachine = 2 //不将消息同步到FromAccount
	msg.MsgRandom = int(time.Now().Unix())
	msg.MsgTimeStamp = int(time.Now().Unix())
	msg.FromAccount = fmt.Sprintf("%d", 100000)
	msg.ToAccount = fmt.Sprintf("%d", uin)
	msg.MsgBody = []im.IMMsgBody{leaveFrozenMsgBody}
	msg.MsgLifeTime = 604800

	var offlinePush im.OfflinePushInfo

	var extInfo im.NotifyExtInfo

	extInfo.NotifyType = constant.ENUM_NOTIFY_TYPE_LEAVE_FROZEN
	extInfo.Content = ""

	se, _ := json.Marshal(extInfo)

	content := "鲁磨路最热路边摊，竟然是它！😱"

	offlinePush.PushFlag = 0
	offlinePush.Desc = content
	offlinePush.Ext = string(se)
	offlinePush.Apns = im.ApnsInfo{1, "", "", ""} //badge不计数
	offlinePush.Ands = im.AndroidInfo{"噗噗"}

	msg.OfflinePush = offlinePush

	return
}

func SendPushMsg(uin int64) (err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Errorf(err.Error())
		return
	}

	sig, err := im.GetAdminSig()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	msg, err := MakeIMPushMsg(uin)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendPushMsgReq uin %d, msg %+v", uin, msg)

	data, err := json.Marshal(&msg)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	url := fmt.Sprintf("https://console.tim.qq.com/v4/openim/sendmsg?usersig=%s&identifier=%s&sdkappid=%d&random=%d&contenttype=json",
		sig, constant.ENUM_IM_IDENTIFIER_ADMIN, constant.ENUM_IM_SDK_APPID, time.Now().Unix())

	hrsp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	body, err := ioutil.ReadAll(hrsp.Body)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	var rsp im.IMSendMsgRsp

	err = json.Unmarshal(body, &rsp)
	if err != nil {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, err.Error())
		log.Errorf(err.Error())
		return
	}

	log.Errorf("SendPushMsgRsp uin %d, rsp %+v", uin, rsp)

	if rsp.ErrorCode != 0 {
		err = rest.NewAPIError(constant.E_IM_REQ_SEND_VOTE_MSG, rsp.ErrorInfo)
		log.Errorf(err.Error())
		return
	}

	return
}
