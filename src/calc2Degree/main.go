package main

import (
	"common/constant"
	"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"common/util"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	confFile string
	log      = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/calc_2degree_svr.conf", "默认配置文件路径")
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &env.Calc2DegreeConfig))
	panicUnless(env.InitLog(env.Calc2DegreeConfig.Log.LogPath, env.Calc2DegreeConfig.Log.LogFileName, env.Calc2DegreeConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.Calc2DegreeConfig.DbInsts))
	panicUnless(myredis.Init(env.Calc2DegreeConfig.RedisInsts, env.Calc2DegreeConfig.RedisApps))

	log.Errorf("start server....")

	calc2DegreeFriend()

	t := time.Tick(time.Second * time.Duration(env.Calc2DegreeConfig.CalcGap))

	for {
		select {
		case <-t:
			calc2DegreeFriend()
		}
	}
}

func DelAllResult() (err error) {

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_2DEGREE_FRIENDS)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("begin scan keys")

	keys, err := app.GetAllKeys()

	if err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Errorf("begin DelAllResult, keys cnt %d", len(keys))

	for key, _ := range keys {
		app.Del(key)
	}

	return
}

func calc2DegreeFriend() (err error) {

	log.Errorf("begin calc2DegreeFriend....")

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_2DEGREE_FRIENDS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//每个人的好友列表
	friends := make(map[int64]map[int64]int)

	//记录唯一UIN列表 方便后面排序
	uniqUins := make(map[int]int)

	s := 0
	e := 100
	total := 0

	sql := fmt.Sprintf(`select count(uin) from friends`)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}

	log.Errorf("friends pair cnt %d", total)

	for {

		sql = fmt.Sprintf(`select uin, friendUin from friends limit %d, %d`, s, e)

		rows, err = inst.Query(sql)
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
			log.Errorf(err.Error())
			return
		}
		defer rows.Close()

		for rows.Next() {

			var uin, friendUin int64
			rows.Scan(&uin, &friendUin)

			uniqUins[int(uin)] = int(uin)

			if _, ok := friends[uin]; !ok {
				friends[uin] = make(map[int64]int)
			}

			if friendUin == uin {
				continue
			}

			friends[uin][friendUin] = 1
		}

		s += 100

		if s >= total {
			break
		}
	}

	log.Errorf("uniq uin cnt %d", len(uniqUins))

	uins := make([]int64, 0)
	//排序
	ps := util.SortMap1(uniqUins)
	for _, v := range ps {
		uins = append(uins, int64(v.Key))
	}

	log.Errorf("sorted uin cnt %d, %+v", len(uins), uins)

	res := make(map[string]int)

	//两重循环
	for i := 0; i < len(uins); i++ {

		uin1 := uins[i]
		//我的好友列表
		mUins1 := friends[uin1]

		cursor := i

		for j := cursor + 1; j < len(uins); j++ {

			uin2 := uins[j]

			keyStr := fmt.Sprintf("%d_%d", uin1, uin2)

			//好友的好友列表
			mUins2 := friends[uin2]

			//计算共同好友数目
			cnt := 0

			sharedUins := make([]int64, 0)

			//t是好友的好友
			for t, _ := range mUins1 {

				//去除掉本身的2个人
				if t == uin1 || t == uin2 {
					continue
				}

				if _, ok := mUins2[t]; ok {
					cnt += 1

					sharedUins = append(sharedUins, t)
				}
			}

			if cnt > 0 {
				log.Errorf("uin1 %d, uin2 %d, sharefriendCnt %d, sharedUins %+v", uin1, uin2, cnt, sharedUins)
				res[keyStr] = cnt
			}
		}
	}

	//删除所有的现有记录
	err = DelAllResult()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for k, v := range res {

		if v == 0 {
			continue
		}

		//log.Debugf("result %s %d", k, v)

		a := strings.Split(k, "_")
		if len(a) != 2 {
			log.Errorf("invalid pairs %s", k)
			continue
		}

		key := strings.TrimSpace(a[0])
		member := strings.TrimSpace(a[1])

		err1 := app.ZAdd(key, int64(v), member)
		if err1 != nil {
			log.Errorf(err1.Error())
		}

		err1 = app.ZAdd(member, int64(v), key)
		if err1 != nil {
			log.Errorf(err1.Error())
		}
	}

	log.Errorf("finished calc2DegreeFriend....")

	return
}

func UpdateByChangeFriendShip(uin1, uin2 int64, op int) (err error) {

	log.Errorf("UpdateByAddFriend begin, uin1 %d, uin2 %d, op %d", uin1, uin2, op)

	if uin1 == uin2 || uin1 == 0 || uin2 == 0 {
		return
	}

	incrCnt := int64(-1)
	if op == 1 {
		incrCnt = 1
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	app, err := myredis.GetApp(constant.ENUM_REDIS_APP_2DEGREE_FRIENDS)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//每个人的好友列表
	friendUins1 := make(map[int64]int)
	friendUins2 := make(map[int64]int)

	sql := fmt.Sprintf(`select uin, friendUin from friends where uin = %d or uin = %d`, uin1, uin2)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Errorf(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uin, friendUin int64
		rows.Scan(&uin, &friendUin)

		if uin == uin1 {
			friendUins1[friendUin] = 1
		} else {
			friendUins2[friendUin] = 1
		}
	}

	//uin1->uin2 : uin2和uin1的每个好友的共同好友数+1
	for uid, _ := range friendUins1 {

		val, err1 := app.ZIncrBy(fmt.Sprintf("%d", uin2), fmt.Sprintf("%d", uid), incrCnt)
		if err1 != nil {
			log.Errorf(err1.Error())
		} else {
			log.Errorf("uin1 %d, uin2 %d, new sharedFriendCnt %d", uin2, uid, val)
		}

		val, err1 = app.ZIncrBy(fmt.Sprintf("%d", uid), fmt.Sprintf("%d", uin2), incrCnt)
		if err1 != nil {
			log.Errorf(err1.Error())
		} else {
			log.Errorf("uin1 %d, uin2 %d, new sharedFriendCnt %d", uid, uin2, val)
		}

	}

	for uid, _ := range friendUins2 {

		val, err1 := app.ZIncrBy(fmt.Sprintf("%d", uin1), fmt.Sprintf("%d", uid), incrCnt)
		if err1 != nil {
			log.Errorf(err1.Error())
		} else {
			log.Errorf("uin1 %d, uin2 %d, new sharedFriendCnt %d", uin1, uid, val)
		}

		val, err1 = app.ZIncrBy(fmt.Sprintf("%d", uid), fmt.Sprintf("%d", uin1), incrCnt)
		if err1 != nil {
			log.Errorf(err1.Error())
		} else {
			log.Errorf("uin1 %d, uin2 %d, new sharedFriendCnt %d", uid, uin1, val)
		}
	}

	log.Errorf("UpdateByAddFriend end, uin1 %d, uin2 %d", uin1, uin2)

	return
}
