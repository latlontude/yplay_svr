package mydb

import (
	"common/env"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var (
	log = env.NewLogger("db")

	insts map[string]*sql.DB
)

func Init(dbInsts map[string]env.DataBase) (err error) {

	insts = make(map[string]*sql.DB)

	for name, config := range dbInsts {

		url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			config.User,
			config.Passwd,
			config.Host,
			config.Port,
			config.DbName,
			"utf8mb4")

		inst, err1 := sql.Open("mysql", url)
		if err1 != nil {
			fmt.Printf("sql open %s error %s\n", url, err1.Error())
			err = err1
			return
		}

		inst.SetMaxOpenConns(config.MaxOpenConn)
		inst.SetMaxIdleConns(config.MaxIdleConn)

		insts[name] = inst
	}

	return
}

func GetInst(name string) (inst *sql.DB) {

	inst, ok := insts[name]
	if ok {
		return
	}

	log.Errorf("db getinst %s return nil", name)

	inst = nil
	return
}

// 事务方式执行多条sql语句(不关心affectedRows)
func Exec(inst *sql.DB, sqlArray []string) error {

	// 开始事务
	tx, err := inst.Begin()
	if err != nil {
		//log.Error(err)
		return err
	}

	for _, sql := range sqlArray {
		if _, err := tx.Exec(sql); err != nil {
			//log.Error("Exec sql ", sql, " failed:", err)
			tx.Rollback()
			return err
		}
		//log.Info("Exec sql ", sql)
	}

	err = tx.Commit()
	if err != nil {
		//log.Error(err)
		return err
	}

	return nil
}
