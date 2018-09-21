package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

//获得点赞数 通用方法
func GetLikeCntByType(likeId int, likeType int) (cnt int, err error) {

	log.Debugf("start get like cnt,likeId:%d,likeType:%d", likeId, likeType)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}
	// 点赞数
	sql := fmt.Sprintf(`select count(id) as cnt from v2likes where type = %d and likeId = %d and likeStatus != 2`, likeType, likeId)

	log.Debugf("likeCntSql:%s", sql)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&cnt)
	}

	log.Debugf("get like count likeType:%d likeId:%d, likeCnt:%d", likeType, likeId, cnt)
	return
}

//判断我是否点过赞
func CheckIsILike(uin int64, likeId int, likeType int) (ret bool, err error) {

	log.Debugf("start check is  like, likeType:%d likeId:%d, uin:%d", likeType, likeId, uin)
	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select id from v2likes where type = %d and likeId = %d and ownerUid = %d and likeStatus != 2`, likeType, likeId, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		rows.Scan(&id)
		ret = true
	}
	log.Debugf("check is  like, likeType:%d likeId:%d, ret:%t", likeType, likeId, ret)
	return
}
