package st

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
)

type UserGemStatInfo struct {
	Uin      int64  `json:"uin"`
	QId      int    `json:"qid"`
	QText    string `json:"qtext"`
	QIconUrl string `json:"qiconUrl"`
	GemCnt   int    `json:"gemCnt"`
}

func (this *UserGemStatInfo) String() string {

	return fmt.Sprintf(`UserGemStatInfo{Uin:%d, QId:%d, QText:%s, QIconUrl:%s, GemCnt:%d}`,
		this.Uin, this.QId, this.QText, this.QIconUrl, this.GemCnt)
}

func IncrGemCnt(uin int64) (err error) {

	if uin == 0 {
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err.Error())
		return
	}

	//更新钻石计数
	res, err := inst.Exec(fmt.Sprintf(`update userStat set statValue = statValue + 1 where uin = %d and statField = %d`, uin, constant.ENUM_USER_STAT_GEM_CNT))
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

	//不存在则插入
	if ra == 0 {
		res, err = inst.Exec(fmt.Sprintf(`insert ignore into userStat values(%d, %d, %d)`, uin, constant.ENUM_USER_STAT_GEM_CNT, 1))
		if err != nil {
			err = rest.NewAPIError(constant.E_DB_EXEC, err.Error())
			log.Error(err.Error())
			return
		}
	}

	return
}
