package common

import (
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"svr/st"
)

func GetV2Reply(replyId int) (reply st.ReplyInfo, ownerUid int64, err error) {
	if replyId == 0 {
		log.Errorf("replyId is zero")
		return
	}

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select replyId,replyContent,fromUid,toUid,replyTs  from v2replys where replyId = %d`, replyId)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	var fromUid int64
	var toUid int64

	for rows.Next() {
		rows.Scan(&reply.ReplyId, &reply.ReplyContent, &fromUid, &toUid, &reply.ReplyTs)
	}

	//被点赞对象
	ownerUid = toUid

	if fromUid > 0 {
		ui, err1 := st.GetUserProfileInfo(fromUid)
		if err1 != nil {
			log.Error(err1.Error())
		}

		reply.ReplyFromUserInfo = ui
	}

	if toUid > 0 {
		ui, err1 := st.GetUserProfileInfo(toUid)
		if err1 != nil {
			log.Error(err1.Error())
		}
		reply.ReplyToUserInfo = ui
	}

	return
}
