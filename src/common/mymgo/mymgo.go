package mymgo

import (
	"common/env"
	"fmt"
	"gopkg.in/mgo.v2"
	"time"
)

var (
	log = env.NewLogger("mgo")

	sessions map[string]*mgo.Session
)

func Init(mgoInsts map[string]env.MgoInst) (err error) {

	sessions = make(map[string]*mgo.Session)

	for name, config := range mgoInsts {

		// dialInfo := &mgo.DialInfo{
		//        Addrs:     []string{config.Hosts},
		//        Direct:    false,
		//        Timeout:   config.Timeout,
		//        Database:  config.DbName,
		//        Source:    "admin",
		//        Username:  config.Username,
		//        Password:  config.Password,
		//        PoolLimit: 4096, //Session.SetPoolLimit
		//    }

		url := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin",
			config.User,
			config.Passwd,
			config.Hosts,
			config.DbName)

		session, err1 := mgo.Dial(url)
		if err1 != nil {
			fmt.Sprintf("mgo open %s error %s", name, err1.Error())
			err = err1
			return
		}

		session.SetMode(mgo.Monotonic, true)
		session.SetSocketTimeout(time.Duration(config.TimeOut) * time.Second)
		session.SetPoolLimit(config.MaxConn)

		sessions[name] = session
	}

	return
}

func GetSession(name string) (inst *mgo.Session) {

	inst, ok := sessions[name]
	if ok {
		return inst.Clone()
	}

	log.Errorf("mgo getinst %s return nil", name)

	inst = nil
	return
}
