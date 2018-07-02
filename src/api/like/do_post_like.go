package like

import (
	"api/v2push"
	"common/constant"
	"common/mydb"
	"common/rest"
	"fmt"
	"net/http"
)

type PostLikeReq struct {
	Uin   int64  `schema:"uin"`
	Token string `schema:"token"`
	Ver   int    `schema:"ver"`

	Qid    int `schema:"qid"`
	LikeId int `schema:"likeId"`
	Typ    int `schema:"type"` //1 给回答点赞，2给评论点赞，3给回应点赞
}

type PostLikeRsp struct {
	Code int `json:"code"`
}

func doPostLike(req *PostLikeReq, r *http.Request) (rsp *PostLikeRsp, err error) {

	log.Debugf("uin %d, PostLikeReq %+v", req.Uin, req)

	code, err := PostLike(req.Uin, req.Qid, req.LikeId, req.Typ)

	if err != nil {
		log.Errorf("uin %d, PostLike error, %s", req.Uin, err.Error())
		return
	}

	rsp = &PostLikeRsp{code}

	log.Debugf("uin %d, PostLikeRsp succ, %+v", req.Uin, rsp)

	return
}

func PostLike(uin int64, qid, likeId, typ int) (code int, err error) {

	if qid == 0 || likeId == 0 || typ == 0 {
		err = rest.NewAPIError(constant.E_INVALID_PARAM, "invalid params")
		log.Error(err)
		return
	}

	code = -1

	inst := mydb.GetInst(constant.ENUM_DB_INST_YPLAY)
	if inst == nil {
		err = rest.NewAPIError(constant.E_DB_INST_NIL, "db inst nil")
		log.Error(err)
		return
	}

	sql := fmt.Sprintf(`select id ,likeStatus from v2likes where qid = %d and type = %d and likeId = %d and ownerUid = %d `, qid, typ, likeId, uin)
	rows, err := inst.Query(sql)
	if err != nil {
		err = rest.NewAPIError(constant.E_DB_QUERY, err.Error())
		log.Error(err)
		return
	}
	defer rows.Close()



	id := 0
	likeStatus := 0
	for rows.Next() {
		rows.Scan(&id,&likeStatus)
	}

	pushAble := false
	log.Debugf("sql:%s  id =%d, likeStatus = %d",sql,id,likeStatus)

	if id == 0 && likeStatus == 0 {
		//insert
		err = insertV2Like(uin,likeId,qid,typ,inst)

		//新增的点赞可以发推送
		pushAble = true
		log.Debugf("insert like id:%d , likeStatus : %d",id,likeStatus)
	}else if id != 0 &&likeStatus == 2 {
		//delete -> update
		err = updateV2LikeById(id,1,inst)
		log.Debugf("update like id:%d , likeStatus : %d",id,likeStatus)

		//取消赞后再点赞 不发推送

	}else {
		code = 0
		log.Debugf("repeat like!")
		return
	}


	code = 0

	if pushAble {
		//回答 评论 回复被点赞 发推送
		if typ == 1 {
			go v2push.SendBeLikedAnswerPush(uin, qid, likeId)
		}else if typ == 2 {
			go v2push.SendBeLikedCommentPush(uin,qid,likeId)
		}else if typ == 3 {
			go v2push.SendBeLikedReplyPush(uin,qid,likeId)
		}
	}


	return
}
