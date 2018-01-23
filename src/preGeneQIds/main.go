package main

import (
	"api/geneqids"
	"common/constant"
	"common/env"
	"common/mydb"
	"common/myredis"
	"common/rest"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	confFile string
	log      = env.NewLogger("main")
)

func init() {
	flag.StringVar(&confFile, "f", "../etc/pre_gene_qids_svr.conf", "默认配置文件路径")
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

	panicUnless(env.InitConfig(confFile, &env.PreGeneQIdsConfig))
	panicUnless(env.InitLog(env.PreGeneQIdsConfig.Log.LogPath, env.PreGeneQIdsConfig.Log.LogFileName, env.PreGeneQIdsConfig.Log.LogLevel))
	panicUnless(mydb.Init(env.PreGeneQIdsConfig.DbInsts))
	panicUnless(myredis.Init(env.PreGeneQIdsConfig.RedisInsts, env.PreGeneQIdsConfig.RedisApps))

	time.Sleep(3 * time.Second)

	log.Errorf("start preGeneQIds....")

	modName := strings.Trim(env.PreGeneQIdsConfig.GeneModInfo.ModName, " ")
	ids := env.PreGeneQIdsConfig.GeneModInfo.Ids

	//include only gene include ids
	//exclude gene alluins except ids
	if modName != "include" && modName != "exclude" {
		log.Errorf("invalid mod name %s", modName)
		return
	}

	uinsStr := strings.Split(ids, ":")
	uids := make([]int64, 0)

	for _, str := range uinsStr {
		uid, _ := strconv.Atoi(str)

		if uid > 0 {
			uids = append(uids, int64(uid))
		}
	}

	log.Errorf("mod name %s, ids %+v", modName, uids)

	if modName == "exclude" {

		uins, err := GetAllUins()
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		t_uids := make([]int64, 0)

		for _, t1 := range uins {

			find := false
			for _, t2 := range uids {
				if t2 == t1 {
					find = true
					break
				}
			}

			if !find {
				t_uids = append(t_uids, t1)
			}
		}

		uids = t_uids
	}

	err := geneqids.GetAllQIds()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	for _, uin := range uids {

		total, err1 := geneqids.Gene(uin)
		if err1 != nil {
			log.Errorf(err.Error())
		} else {
			log.Errorf("uin %d, pregeneqids total %d", uin, total)
		}
	}

	log.Errorf("end preGeneQIdsr....")

	time.Sleep(3 * time.Second)
}

func GetAllUins() (uins []int64, err error) {

	uins = make([]int64, 0)

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Errorf(err.Error())
		return
	}

	sql := fmt.Sprintf(`select uin, gender, grade, schoolId, schoolType from profiles`)

	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {

		var uin int64
		var gender int
		var grade int
		var schoolId int
		var schoolType int

		rows.Scan(&uin, &gender, &grade, &schoolId, &schoolType)

		//if uin > 0 && gender > 0 && schoolId > 0 {
		uins = append(uins, uin)
		//}
	}

	return
}
