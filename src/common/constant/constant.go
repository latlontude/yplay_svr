package constant

const DEFAULT_PAGE_SIZE = 10

const ENUM_QUESTION_BATCH_SIZE = 12 //每次批量拉取题目数
const ENUM_OPTION_BATCH_SIZE = 4    //每个题目下的选择项数目

const ENUM_ADDR_MAX_BATCH_UPLOAD_SIZE = 100 //通讯录每次上传的最大数目

const DEFAULT_MAX_SEND_SMS_CNT = 5 // 投票给未注册好友，最大发送短信次数5次

const TOKEN_YPLAY_AES_KEY = "frankshi;@yeejay"

const ENUM_USER_STAT_GEM_CNT = 1    //钻石统计
const ENUM_USER_STAT_FRIEND_CNT = 2 //好友统计

const ENUM_USER_GRADE_1 = 1
const ENUM_USER_GRADE_2 = 2
const ENUM_USER_GRADE_3 = 3
const ENUM_USER_GRADE_4 = 4
const ENUM_USER_GRADE_5 = 5
const ENUM_USER_GRADE_GRADUATE = 100

const ENUM_SHOOL_TYPE_JUNIOR = 1      //初中
const ENUM_SHOOL_TYPE_HIGH = 2        //高中
const ENUM_SCHOOL_TYPE_UNIVERSITY = 3 //大学

const ENUM_DB_INST_YPLAY = "yplay"
const ENUM_MGO_INST_YPLAY = "yplay"

const ENUM_REDIS_APP_SMS = "sms"                                           //短信验证码
const ENUM_REDIS_APP_TOKEN = "token"                                       //登录token
const ENUM_REDIS_APP_LAST_READ_ADDFRIEND_MSG_ID = "lastreadaddfriendmsgid" //用户上一次读取的加好友消息时间
const ENUM_REDIS_APP_FEED_MSG = "feedmsg"                                  //用户的新的动态消息
//const ENUM_REDIS_APP_VOTED_QIDS = "votedqids"                              //用户已经投票的问题类别
//const ENUM_REDIS_APP_QUESTION_CURSOR = "qcursor"                           //用户下发题目的游标，每次循环完成后，再从头开始
//const ENUM_REDIS_APP_USER_VOTE_PROGRESS = "voteprogress"                   //用户当前未回答的题目
const ENUM_REDIS_APP_2DEGREE_FRIENDS = "2degreefriends"   //2度好友的统计
const ENUM_REDIS_APP_LAST_READ_FEED_MS = "lastreadfeedms" //最近一次读取feed的毫秒时间
//const ENUM_REDIS_APP_IMGROUP            = "imgroup"        //IM创建群 每次投票对应一个群
const ENUM_REDIS_APP_LAST_QINFO = "lastqinfo"
const ENUM_REDIS_APP_USER_QID_VOTED_CNT = "userqidvotedcnt"
const ENUM_REDIS_APP_INVITE_CODE = "invitecode"
const ENUM_REDIS_APP_SUBMIT_LAST_READ_ONLINE_TS = "submitlastreadonline" //上一次读取已经上线的投稿时间,用于判断哪些是新上线的标志
const ENUM_REDIS_APP_USER_PV_CNT = "userpvcnt"
const ENUM_REDIS_APP_PRE_GENE_QIDS = "pregeneqids"
const ENUM_REDIS_APP_SNAPCHAT_SESSION = "snapchatsession"
const ENUM_REDIS_APP_VOTECHAT_SESSION = "votechatsession"
const ENUM_REDIS_APP_STORY_MSG = "storymsg"                       //存储每条story的信息
const ENUM_REDIS_APP_STORY_STAT = "storystat"                     //存储每条story的观看记录
const ENUM_REDIS_APP_FRIEND_STORY_LIST = "friendstorymsg"         //每个人的朋友圈story列表
const ENUM_REDIS_APP_MY_STORY_LIST = "mystorymsg"                 //自己的story列表
const ENUM_REDIS_APP_LAST_READ_STORY_MS = "lastreadstoryms"       //最近一次读story的时间
const ENUM_REDIS_APP_USER_LOOKED_OPTION_UINS = "lookedoptionuins" //用户答题中看到的选项UIN列表用于优化选项
const ENUM_REDIS_APP_INVITE_FRIEND_BY_VOTE = "invitedcnt"         //记录投票给未注册好友每日发送短信的次数

const ENUM_VOTE_STATUS_INIT = 0
const ENUM_VOTE_STATUS_REPLY = 1
const ENUM_VOTE_STATUS_REPLY_REPLY = 2


// add friends

const ENUM_ADD_FRIEND_STATUS_INIT = 0           //
const ENUM_ADD_FRIEND_STATUS_ACCEPT = 1         //接受
const ENUM_ADD_FRIEND_STATUS_IGNORE = 2         //拒绝

const ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_REGISTED = 1         //通讯录非好友
const ENUM_RECOMMEND_FRIEND_TYPE_ADDR_BOOK_NOT_REGISTED = 2     //通讯录未注册
const ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL = 3                //同校所有
const ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GRADE = 4          //同校同年级
const ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_BOY = 5            //同校男生
const ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_GIRL = 6           //同校女生
const ENUM_RECOMMEND_FRIEND_TYPE_2DEGREE_FRIEND = 7             //共同好友
const ENUM_RECOMMEND_FRIEND_SEARCH = 8                          //搜索
const ENUM_RECOMMEND_FRIEND_TYPE_SAME_SCHOOL_DEPT = 9           //大学同学院

const ENUM_SNS_STATUS_NOT_FRIEND = 0                            //非好友
const ENUM_SNS_STATUS_IS_FRIEND = 1                             //好友
const ENUM_SNS_STATUS_HAS_INVAITE_FRIEND = 2                    //单向加好友
const ENUM_SNS_STATUS_FRIEND_HAS_INVAITE_ME = 3                 //单向加好友
const ENUM_SNS_STATUS_NOT_INVITE_BY_SMS = 4
const ENUM_SNS_STATUS_HAS_INVAITE_BY_SMS = 5

const ENUM_DEVICE_UUID_MIN = 1500000000000

const ENUM_IM_IDENTIFIER_ADMIN = "frankshi"
const ENUM_IM_SDK_APPID = 1400046572

//const ENUM_PROFILE_MOD_MAX_CNT = 2
//const ENUM_PROFILE_GENDER_MOD_MAX_CNT = 1


//用户信息修改
const ENUM_PROFILE_MOD_FIELD_MIN = 0
const ENUM_PROFILE_MOD_FIELD_NICKNAME = 1
const ENUM_PROFILE_MOD_FIELD_USERNAME = 2
const ENUM_PROFILE_MOD_FIELD_SCHOOLGRADE = 3
const ENUM_PROFILE_MOD_FIELD_GENDER = 4
const ENUM_PROFILE_MOD_FIELD_AGE = 5
const ENUM_PROFILE_MOD_FIELD_COUNTRY = 6
const ENUM_PROFILE_MOD_FIELD_PROVINCE = 7
const ENUM_PROFILE_MOD_FIELD_CITY = 8
const ENUM_PROFILE_MOD_FIELD_HOMETOWN = 9
const ENUM_PROFILE_MOD_FIELD_MAX = 10

//const ENUM_FREEZE_SECONDS        = 60*3600

const ENUM_FREEZE_STATUS_FROZEND = 1
const ENUM_FREEZE_STATUS_NOT_FROZEND = 0

const ENUM_NOTIFY_TYPE_LEAVE_FROZEN = 1
const ENUM_NOTIFY_TYPE_IM = 2
const ENUM_NOTIFY_TYPE_ADD_FRIEND = 3
const ENUM_NOTIFY_TYPE_SUBMIT_ADD_NEW_HOT = 4


//const label for qid
const EXPERIENCE_QID_LABEL_COUNT = 1    //每个问题最多 多少个标签
