package sns

import (
	"api/im"
	"api/story"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
	"time"
)

type RemoveFriendS2SReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	ToUin int64 `schema:"toUin"`
}

type RemoveFriendS2SRsp struct {

}

func doRemoveFriendS2S(req *RemoveFriendS2SReq, r *http.Request) (rsp *RemoveFriendS2SRsp, err error) {
	log.Errorf("uin %d, RemoveFriendReq %+v", req.Uin, req)

	err = RemoveFriendS2S(req.Uin, req.ToUin)
	if err != nil {
		log.Errorf("uin %d, RemoveFriendRsp error, %s", req.Uin, err.Error())
		return
	}

	rsp = &RemoveFriendS2SRsp{}

	log.Errorf("uin %d, RemoveFriendRsp succ, %+v", req.Uin, rsp)

	return
}

func RemoveFriendS2S(uin, friendUin int64) (err error) {

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select status from friends where uin = %d and friendUin = %d`, uin, friendUin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	find := false
	var status int

	for rows.Next() {
		rows.Scan(&status)
		find = true
	}

	//已经不是好友，直接返回成功
	if !find {
		//err = rest.NewAPIError(constant.E_PERMI_DENY, "permission denied")
		//log.Error(err.Error())
		return
	}

	ts := time.Now().Unix()
	ms := time.Now().UnixNano() / 1000000

	sqls := make([]string, 0)
	sqls = append(sqls, fmt.Sprintf(`delete from friends where uin = %d and friendUin = %d`, uin, friendUin))
	sqls = append(sqls, fmt.Sprintf(`delete from friends where uin = %d and friendUin = %d`, friendUin, uin))
	//更新我的好友列表的版本号
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, friendUin, ms, ts, ms, ts))
	sqls = append(sqls, fmt.Sprintf(`insert into friendListVer values(%d, %d, %d) on duplicate key update ver = %d, ts = %d`, uin, ms, ts, ms, ts))

	//解除好友之后把之前的添加好友消息都清理掉
	sqls = append(sqls, fmt.Sprintf(`delete from addFriendMsg where fromUin = %d and toUin = %d`, uin, friendUin))
	sqls = append(sqls, fmt.Sprintf(`delete from addFriendMsg where fromUin = %d and toUin = %d`, friendUin, uin))

	//解除好友之后，好友计数要减少
	sqls = append(sqls, fmt.Sprintf(`update userStat set statValue = statValue -1 where uin = %d and statField = %d and statValue > 0`, uin, constant.ENUM_USER_STAT_FRIEND_CNT))
	sqls = append(sqls, fmt.Sprintf(`update userStat set statValue = statValue -1 where uin = %d and statField = %d and statValue > 0`, friendUin, constant.ENUM_USER_STAT_FRIEND_CNT))

	err = mydb.Exec(inst, sqls)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
		log.Error(err.Error())
		return
	}

	//更新好友的好友关系
	//.......
	go im.RemoveFriendS2S(uin, friendUin)

	//从双方的24小时新闻中删除对方发表的新闻，以及双方观看对方的新闻的记录
	go story.RemoveStoryByDelFriend(uin, friendUin)
	go story.RemoveStoryByDelFriend(friendUin, uin)

	return
}
