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
	"time"
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
	flag.StringVar(&confFile, "f", "gene_snap_sessions.conf", "默认配置文件路径")
}

type UserPair struct {
	Uin1 int64
	Uin2 int64
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

	fmt.Printf("users cnt %d\n", len(users))

	panicUnless(env.InitConfig(confFile, &env.Config))
	panicUnless(env.InitLog(env.Config.Log.LogPath, env.Config.Log.LogFileName, env.Config.Log.LogLevel))
	panicUnless(myredis.Init(env.Config.RedisInsts, env.Config.RedisApps))

	for _, pair := range users {

		uin1 := pair.Uin1
		uin2 := pair.Uin2

		log.Errorf("begin uin1 %d, uin2 %d, create snap chat session", uin1, uin2)

		if uin1 == uin2 || uin1 <= 0 || uin2 <= 0 {
			log.Errorf("invalid uin1 %d, uin2 %d", uin1, uin2)
			continue
		}

		groupId, err := im.CreateSnapChatSesson(uin1, uin2)
		if err != nil {
			log.Errorf("uin1 %d, uin2 %d, create snap chat session error %s", uin1, uin2, err.Error())
			continue
		}

		log.Errorf("uin1 %d, uin2 %d, create snap chat session success, groupId %s", uin1, uin2, groupId)
	}

	log.Errorf("end geneSnapSessions....")

	time.Sleep(3 * time.Second)

	return
}

func GetUins(fileName string) (users []UserPair, err error) {

	f, err := os.Open(fileName)
	if err != nil {
		return
	}

	users = make([]UserPair, 0)

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

		a := strings.Split(line, " ")
		if len(a) != 2 {
			err = errors.New("invalid line input, " + line)
			return
		}

		u1, _ := strconv.Atoi(a[0])
		u2, _ := strconv.Atoi(a[1])

		users = append(users, UserPair{int64(u1), int64(u2)})
	}

	return
}
