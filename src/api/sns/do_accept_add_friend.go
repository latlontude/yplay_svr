package sns

import (
	"common/constant"
	//"common/env"
	"common/rest"
	//"common/myredis"
	"api/geneqids"
	"api/im"
	"common/mydb"
	"fmt"
	"net/http"
	"svr/st"
	"time"
)

type AcceptAddFriendReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	MsgId int64 `schema:"msgId"`
	Act   int   `schema:"act"`
}

type AcceptAddFriendRsp struct {
}

func doAcceptAddFriend(req *AcceptAddFriendReq, r *http.Request) (rsp *AcceptAddFriendRsp, err error) {

	log.Errorf("uin %d, AcceptAddFriendReq %+v", req.Uin, req)

	err = AcceptAddFriend(req.Uin, req.MsgId, req.Act)
	if err != nil {
		log.Errorf("uin %d, AcceptAddFriendRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &AcceptAddFriendRsp{}

	log.Errorf("uin %d, AcceptAddFriendRsp succ, %+v", req.Uin, rsp)

	return
}

func AcceptAddFriend(uin int64, msgId int64, act int) (err error) {

	if uin == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid uin")
		log.Error(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//检查是否存在这样的消息
	sql := fmt.Sprintf(`select fromUin, toUin from addFriendMsg where msgId = %d and toUin = %d`, msgId, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false
	var toUin, fromUin int64

	for rows.Next() {
		rows.Scan(&fromUin, &toUin)
		find = true
	}

	//不存在 则拒绝操作
	if !find {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "permission denied")
		log.Error(err.Error())
		return
	}

	//不能加自己为好友
	if fromUin == toUin {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "add self friend")
		log.Error(err.Error())
		return
	}

	//用户ID非法
	if fromUin == 0 || toUin == 0 {
		err = rest.NewAPIError(constant.E_PERMI_DENY, "uin zero!")
		log.Error(err.Error())
		return
	}

	//更新状态
	ts := time.Now().Unix()
	ms := time.Now().UnixNano() / 1000000 //毫秒

	status := constant.ENUM_ADD_FRIEND_STATUS_ACCEPT
	//接受加好友请求 删掉原有消息
	//sql = fmt.Sprintf(`delete from addFriendMsg where msgId = %d`, msgId)

	//接受加好友请求，保留原有消息
	sql = fmt.Sprintf(`update addFriendMsg set status = %d, mts = %d where msgId = %d`, status, ts, msgId)

	//忽略加好友请求
	if act > 0 {
		status = constant.ENUM_ADD_FRIEND_STATUS_IGNORE
		sql = fmt.Sprintf(`update addFriendMsg set status = %d, mts = %d where msgId = %d`, status, ts, msgId)
	}

	_, err = inst.Exec(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//如果是忽略请求 则不用做任何事情了......
	if act > 0 {
		return
	}

	//检查两人是否已经是好友，如果是则计数不用再增加
	sql = fmt.Sprintf(`select uin from friends where uin = %d and friendUin = %d`, fromUin, toUin)
	rows, err = inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find = false

	for rows.Next() {
		var tmp int
		rows.Scan(&tmp)
		find = true
	}

	//如果已经是好友，则直接返回，计数和插入操作免去
	if find {
		return
	}

	//添加到我的好友列表中
	sqls := make([]string, 0)
	sqls = append(sqls, fmt.Sprintf(`insert ignore into friends values(%d, %d, %d, %d)`, toUin, fromUin, 0, ts))
	sqls = append(sqls, fmt.Sprintf(`insert ignore into friends values(%d, %d, %d, %d)`, fromUin, toUin, 0, ts))
	//更新我的好友列表的版本号
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, fromUin, ms, ts, ms, ts))
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, toUin, ms, ts, ms, ts))

	err = mydb.Exec(inst, sqls)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	go IncrFriendCnt(fromUin, toUin)

	//如果申请加好友，就不在共同好友出现
	go Fix2DegreeResult(fromUin, toUin)

	//更新好友的好友关系
	go im.SendAcceptAddFriendMsg(fromUin, toUin)

	//go CreateNewSnapSession(fromUin, toUin)
	go CreateSnapSessionAndSendPushMsg(fromUin, toUin)

	go JudgeNeedGeneQids(fromUin, toUin)

	return
}

//如果好友不足四个到超过四个
func JudgeNeedGeneQids(fromUin int64, toUin int64) (err error) {

	uins1, err := st.GetMyFriendUins(fromUin)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		if len(uins1) == 4 || len(uins1) == 8 {
			geneqids.Gene(fromUin)
		}
	}

	uins2, err := st.GetMyFriendUins(toUin)
	if err != nil {
		log.Errorf(err.Error())
	} else {

		if len(uins2) == 4 || len(uins2) == 8 {
			geneqids.Gene(toUin)
		}
	}

	return
}

func IncrFriendCnt(uin, friendUin int64) (err error) {

	if uin == 0 || friendUin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	res, err := inst.Exec(fmt.Sprintf(`update userStat set statValue = statValue + 1 where uin = %d and statField = %d`, uin, constant.ENUM_USER_STAT_FRIEND_CNT))
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	ra, err := res.RowsAffected()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	if ra == 0 {
		res, err = inst.Exec(fmt.Sprintf(`insert ignore into userStat values(%d, %d, %d)`, uin, constant.ENUM_USER_STAT_FRIEND_CNT, 1))
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}
	}

	res, err = inst.Exec(fmt.Sprintf(`update userStat set statValue = statValue + 1 where uin = %d and statField = %d`, friendUin, constant.ENUM_USER_STAT_FRIEND_CNT))
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	ra, err = res.RowsAffected()
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	if ra == 0 {
		res, err = inst.Exec(fmt.Sprintf(`insert ignore into userStat values(%d, %d, %d)`, friendUin, constant.ENUM_USER_STAT_FRIEND_CNT, 1))
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}
	}

	return
}

//fromUin -> toUin
func CreateSnapSessionAndSendPushMsg(uin, uid int64) (err error) {

	//创建固定会话
	_, err = CreateNewSnapSession(uin, uid)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//查找对方的昵称
	sql := fmt.Sprintf(`select nickName from profiles where uin =  %d`, uid)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	nickName := ""
	for rows.Next() {
		rows.Scan(&nickName)
		break
	}

	//发送push
	//主动发起的一方将收到push
	content := fmt.Sprintf("%s: 我们已成为好友，可以互相评价了！ᕕ( ᐛ )ᕗ", nickName)
	im.SendLeaveFrozenMsg(uin, content)

	// if err != nil {
	// 	log.Errorf(err.Error())
	// 	return
	// }

	//模拟发送一条消息 从uid->uin
	//目标是自动在双方的最近连续人加上对方，测试发现在主动发消息的一方不会产生这个，
	//在接受消息的一方会获取不到昵称和头像 先去掉这个逻辑，从客户端逻辑入手
	//im.SendTextMsg(gid, "我们已成为好友啦，开始聊天吧ᕕ( ᐛ )ᕗ", uid, uin)

	return
}

func CreateNewSnapSession(uin, uid int64) (groupId string, err error) {

	log.Errorf("begin uin %d, uid %d, CreateNewSnapSession ", uin, uid)

	if uin == uid || uin <= 0 || uid <= 0 {
		log.Errorf("invalid uin %d, uid %d", uin, uid)
		return
	}

	groupId, err = im.CreateSnapChatSesson(uin, uid)
	if err != nil {
		log.Errorf("uin %d, uid %d, create snap chat session error %s", uin, uid, err.Error())
		return
	}

	log.Errorf("uin1 %d, uin2 %d, create snap chat session success, groupId %s", uin, uid, groupId)

	return
}
