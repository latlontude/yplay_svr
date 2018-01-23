package main

import (
	"api/im"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/rest"
	"container/list"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	confFile string
	log      = env.NewLogger("main")

	TASKS *list.List

	TASK_UINS map[int64]int
)

type Task struct {
	Uin int64
	Ts  int
}

func (this *Task) String() string {
	return fmt.Sprintf(`Task{Uin:%d, Ts:%d}`, this.Uin, this.Ts)
}

func init() {
	flag.StringVar(&confFile, "f", "../etc/frozen_monitor_svr.conf", "默认配置文件路径")
}

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func main() {

	TASKS = list.New()

	TASK_UINS = make(map[int64]int)

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	panicUnless(env.InitConfig(confFile, &env.FrozenMonitorConfig))
	panicUnless(env.InitLog(env.FrozenMonitorConfig.Log.LogPath, env.FrozenMonitorConfig.Log.LogFileName, env.FrozenMonitorConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.FrozenMonitorConfig.DbInsts))

	log.Errorf("start server....")

	MonitorDB()

	t1 := time.Tick(time.Second * time.Duration(env.FrozenMonitorConfig.MonitorDBGap))
	t2 := time.Tick(time.Second * time.Duration(env.FrozenMonitorConfig.MonitorTaskGap))

	for {
		select {
		case <-t1:
			MonitorDB()

		case <-t2:
			MonitorTask()
		}
	}
}

func MonitorDB() (err error) {

	taskStrs := ""
	for e := TASKS.Front(); e != nil; e = e.Next() {
		tmp := e.Value.(Task)
		taskStrs += tmp.String()
	}

	log.Errorf(`begin monitor db, taskCnt %d, tasks %s`, TASKS.Len(), taskStrs)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, freezeTs from freezingStatus where freezeTs > 0 order by freezeTs`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uin int64
		var freezeTs int
		rows.Scan(&uin, &freezeTs)

		//没有处于冷冻状态
		if freezeTs == 0 {
			continue
		}

		//已经存在
		if _, ok := TASK_UINS[uin]; ok {
			continue
		}

		t := Task{uin, freezeTs}

		TASKS.PushBack(t)

		TASK_UINS[uin] = 1

		log.Errorf(`monitor db, task push %+v`, t)
	}

	log.Errorf(`end monitor db, taskCnt %d`, TASKS.Len())

	return
}

func MonitorTask() {

	freezeDuration := env.FrozenMonitorConfig.FreezeDuration

	//毫秒
	now := int(time.Now().Unix())

	for {

		if TASKS.Len() == 0 {
			break
		}

		e := TASKS.Front()

		t := TASKS.Front().Value.(Task)

		ts := t.Ts

		//提前3秒
		if now-ts >= freezeDuration {

			log.Errorf(`monitor task, task pop %+v`, t)

			TASKS.Remove(e)

			delete(TASK_UINS, t.Uin)

			go SendFreezePush(t.Uin, t.Ts)

		} else {
			break
		}
	}

}

func SendFreezePush(uin int64, freezeTs int) (err error) {

	freezeDuration := env.FrozenMonitorConfig.FreezeDuration

	log.Errorf("uin %d, freezeTs %d, check need send frozen leaving push", uin, freezeTs)

	now := int(time.Now().Unix())

	//一定满足的条件: now-freezeTs >= freezeDuration
	//如果间隔的时间太长 就不发push请求
	if now-freezeTs <= 2*freezeDuration {

		log.Errorf("uin %d, will send push msg right now!", uin)

		err = im.SendLeaveFrozenMsg(uin)
		if err != nil {
			log.Error(err.Error())
		}
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	sql := fmt.Sprintf(`update freezingStatus set freezeTs = 0, ts = %d where uin = %d`, now, uin)
	_, err = inst.Exec(sql)
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
