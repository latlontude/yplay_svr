package main

import (
	"api/im"
	"bufio"
	"common/env"
	"common/myredis"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var (
	idFile   string
	confFile string
	log      = env.NewLogger("genesnapchatsession")
)

func panicUnless(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func init() {
	flag.StringVar(&idFile, "i", "user.lst", "用户列表文件")
	flag.StringVar(&confFile, "f", "yplay_svr.conf", "默认配置文件路径")
}

func main() {

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(idFile) == 0 || len(confFile) == 0 {
		fmt.Printf("invalid file [%s], [%s]\n", idFile, confFile)
		return
	}

	users, err := GetUins(idFile)
	if err != nil {
		fmt.Printf("read users from file[%s] error[%s]\n", idFile, err.Error())
		return
	}

	panicUnless(env.InitConfig(confFile, &env.Config))
	panicUnless(env.InitLog(env.Config.Log.LogPath, env.Config.Log.LogFileName, env.Config.Log.LogLevel))
	panicUnless(myredis.Init(env.Config.RedisInsts, env.Config.RedisApps))

	for uin1, uin2 := range users {

		if uin1 == uin2 || uin1 <= 0 || uin2 <= 0 {
			log.Errorf("invalid uin1 %d, uin2 %d", uin1, uin2)
			continue
		}

		groupId, err := im.CreateSnapChatSesson(uin1, uin2)
		if err != nil {
			log.Errorf("create snap chat session error %s", err.Error())
			continue
		}

		log.Errorf("create snap chat session success %s", groupId)
	}

	return
}

func GetUins(fileName string) (users map[int64]int64, err error) {

	f, err := os.Open(fileName)
	if err != nil {
		return
	}

	users = make(map[int64]int64)

	buf := bufio.NewReader(f)
	for {

		line, err1 := buf.ReadString('\n')

		if err1 != nil {
			if err1 == io.EOF {
				err = nil
			} else {
				err = err1
			}
			return
		}

		line = strings.TrimSpace(line)

		a := strings.Split(" ", line)
		if len(a) != 2 {
			err = errors.New("invalid line input, " + line)
			return
		}

		u1, _ := strconv.Atoi(a[0])
		u2, _ := strconv.Atoi(a[1])

		users[int64(u1)] = int64(u2)
	}

	return
}
