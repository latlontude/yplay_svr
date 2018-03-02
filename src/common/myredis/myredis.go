package myredis

import (
	"common/constant"
	"common/env"
	"common/rest"
	"fmt"
	"github.com/garyburd/redigo/redis"
	//"time"
	"errors"
	"strconv"
)

var (
	log = env.NewLogger("redis")

	insts map[string]*RedisInst //instname -> inst
	apps  map[string]*RedisApp  //appname  -> appid+instname
)

type RedisInst struct {
	Pool *redis.Pool
}

type RedisApp struct {
	InstName string
	AppId    string
}

func Init(redisInsts map[string]env.RedisInst, redisApps map[string]env.RedisApp) (err error) {

	insts = make(map[string]*RedisInst)

	for name, config := range redisInsts {

		addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
		pool := newPool(addr, config.Passwd, config.MaxOpenConn, config.MaxIdleConn)

		if pool == nil {
			err = errors.New(fmt.Sprintf("init redis[%s:%d] pool error", config.Host, config.Port))
			return
		}

		insts[name] = &RedisInst{pool}
	}

	apps = make(map[string]*RedisApp)
	for name, config := range redisApps {

		if len(name) == 0 {
			log.Errorf("redis app name empty")
			continue
		}

		if len(config.InstName) == 0 || len(config.AppId) == 0 {
			log.Errorf("redis app's inst or appId empty")
			continue
		}

		apps[name] = &RedisApp{config.InstName, config.AppId}
	}

	return
}

func newPool(addr string, passwd string, maxConn int, maxIdleConn int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdleConn,
		MaxActive:   maxConn, //最大连接数无限制
		IdleTimeout: 0,       //连接不会超时
		Wait:        false,
		Dial: func() (c redis.Conn, err error) {
			c, err = redis.Dial("tcp", addr)

			if err != nil {
				log.Errorf("redis DIAL error %s", err.Error())
				return nil, err
			}

			if _, err = c.Do("AUTH", passwd); err != nil {
				log.Errorf("redis AUTH error %s", err.Error())
				c.Close()
				return nil, err
			}

			return
		},
	}
}

func GetApp(name string) (app *RedisApp, err error) {
	app, ok := apps[name]

	if ok {
		return
	}

	err = rest.NewAPIError(constant.E_REDIS_APP_NOT_FOUND, "redis appname not found")
	return
}

func GetInst(name string) (inst *RedisInst, err error) {

	inst, ok := insts[name]

	if ok {
		return
	}

	err = rest.NewAPIError(constant.E_REDIS_INST_NOT_FOUND, "redis instname not found")
	return
}

func (this *RedisApp) GetConn() (c redis.Conn, err error) {

	inst, err := GetInst(this.InstName)
	if err != nil {
		return
	}

	if inst == nil {
		err = rest.NewAPIError(constant.E_REDIS_INST_NOT_FOUND, "redis inst nil")
		return
	}

	c = inst.Pool.Get()
	if c == nil {
		err = rest.NewAPIError(constant.E_REDIS_GET_CONN, "conn nil")
		return
	}

	return
}

func (this *RedisApp) Get(key string) (val string, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	val, err = redis.String(conn.Do("get", rkey))
	if err != nil {

		if err == redis.ErrNil {

			err = rest.NewAPIError(constant.E_REDIS_KEY_NO_EXIST, "key not exist,"+err.Error())
			return

		} else {
			err = rest.NewAPIError(constant.E_REDIS_GET, "get error,"+err.Error())
			log.Errorf(err.Error())
			return
		}
	}

	return
}

func (this *RedisApp) GetAllKeys() (keys map[string]int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	keys = make(map[string]int)

	var items []string
	cursor := 0

	for {

		params := make([]interface{}, 0)

		params = append(params, fmt.Sprintf("%d", cursor))
		params = append(params, fmt.Sprintf("%s", "count"))
		params = append(params, fmt.Sprintf("%d", 100))
		params = append(params, fmt.Sprintf("%s", "match"))
		params = append(params, fmt.Sprintf("%s_*", this.AppId))

		//fmt.Sprintf("scan %d count 100 match %s_*", cursor, this.AppId)

		vals, err1 := redis.Values(conn.Do("scan", params...))
		if err1 != nil {

			if err1 == redis.ErrNil {

				err = rest.NewAPIError(constant.E_REDIS_KEY_NO_EXIST, "key not exist,"+err1.Error())

			} else {
				err = rest.NewAPIError(constant.E_REDIS_SCAN, "scan error,"+err1.Error())
				log.Errorf(err.Error())
			}

			return
		}

		vals, err = redis.Scan(vals, &cursor, &items)
		if err != nil {
			err = rest.NewAPIError(constant.E_REDIS_SCAN, "scan error,"+err.Error())
			log.Errorf(err.Error())
			return
		}

		for _, key := range items {
			keys[key] = 1
		}

		if cursor == 0 {
			break
		}
	}

	return
}

func (this *RedisApp) MGet(keys []string) (vals map[string]string, err error) {

	if len(keys) == 0 {
		return
	}

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	i_keys := make([]interface{}, len(keys))

	for i, k := range keys {
		i_keys[i] = fmt.Sprintf("%s_%s", this.AppId, k)
	}

	replys, err := redis.Values(conn.Do("mget", i_keys...))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_MGET, "mulget error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	scan_dsts := make([]string, len(keys))

	err = redis.ScanSlice(replys, &scan_dsts)
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_MGET, "mulget scan error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	vals = make(map[string]string)

	for i, reply := range replys {
		if reply == nil {
			continue
		}

		vals[keys[i]] = scan_dsts[i]
	}

	return
}

func (this *RedisApp) Set(key, val string) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.String(conn.Do("set", rkey, val))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_SET, "set error "+err.Error())
		return
	}

	if reply != "OK" {
		err = rest.NewAPIError(constant.E_REDIS_SET, "set error "+reply)
		return
	}

	return
}

func (this *RedisApp) SetEx(key string, val string, ttl uint32) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.String(conn.Do("setex", rkey, ttl, val))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_SETEX, "setex error"+err.Error())
		return
	}

	//log.Errorf("setex key %s, value %s, ttl %d, return %s", rkey, val, ttl, reply)

	if reply != "OK" {
		err = rest.NewAPIError(constant.E_REDIS_SETEX, "setex error"+reply)
		return
	}

	return
}

func (this *RedisApp) Pexpire(key string, ttl uint32) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.String(conn.Do("pexpire", rkey, ttl))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_PEXPIRE, "pexpire error"+err.Error())
		return
	}

	if reply != "OK" {
		err = rest.NewAPIError(constant.E_REDIS_PEXPIRE, "pexpire error"+reply)
		return
	}

	return
}

func (this *RedisApp) Exist(key string) (exist bool, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.Int(conn.Do("exists", rkey))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_EXIST, "exist error"+err.Error())
		return
	}

	exist = false

	if reply == 1 {
		exist = true
	}

	return
}

func (this *RedisApp) Del(key string) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.Int(conn.Do("del", rkey))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_DEL, "del error"+err.Error())
		return
	}

	if reply != 1 {
		//err = rest.NewAPIError(constant.E_REDIS_DEL, "del return not 1, return "+fmt.Sprintf("%d", reply))
		return
	}

	return
}

func (this *RedisApp) Incr(key string) (val int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.Int(conn.Do("incr", rkey))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_INCR, "incr error "+err.Error())
		return
	}

	val = reply

	return
}

func (this *RedisApp) Decr(key string) (val int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.Int(conn.Do("decr", rkey))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_DECR, "decr error"+err.Error())
		return
	}

	val = reply

	return
}

func (this *RedisApp) ZRange(key string, start, stop int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	vals, err = redis.Strings(conn.Do("ZRANGE", rkey, start, stop))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "range error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRevRange(key string, start, stop int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	vals, err = redis.Strings(conn.Do("ZREVRANGE", rkey, start, stop))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "revrange error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRangeWithScores(key string, start, stop int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	vals, err = redis.Strings(conn.Do("ZRANGE", rkey, start, stop, "WITHSCORES"))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "range error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRevRangeWithScores(key string, start, stop int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	vals, err = redis.Strings(conn.Do("ZREVRANGE", rkey, start, stop, "WITHSCORES"))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "revrange error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRangeByScore(key string, min, max int64, offset, count int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	maxStr := fmt.Sprintf("%d", max)
	if max == -1 {
		maxStr = "+inf"
	}

	minStr := fmt.Sprintf("%d", min)
	if min == -1 {
		minStr = "-inf"
	}

	vals, err = redis.Strings(conn.Do("ZRANGEBYSCORE", rkey, minStr, maxStr, "limit", offset, count))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "range error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRangeByScoreWithScores(key string, min, max int64, offset, count int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	maxStr := fmt.Sprintf("%d", max)
	if max == -1 {
		maxStr = "+inf"
	}

	minStr := fmt.Sprintf("%d", min)
	if min == -1 {
		minStr = "-inf"
	}

	vals, err = redis.Strings(conn.Do("ZRANGEBYSCORE", rkey, minStr, maxStr, "withscores", "limit", offset, count))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "range error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRevRangeByScore(key string, max, min int64, offset, count int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	maxStr := fmt.Sprintf("%d", max)
	if max == -1 {
		maxStr = "+inf"
	}

	minStr := fmt.Sprintf("%d", min)
	if min == -1 {
		minStr = "-inf"
	}

	vals, err = redis.Strings(conn.Do("ZREVRANGEBYSCORE", rkey, maxStr, minStr, "limit", offset, count))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "range error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRevRangeByScoreWithScores(key string, max, min int64, offset, count int) (vals []string, err error) {

	vals = make([]string, 0)

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	maxStr := fmt.Sprintf("%d", max)
	if max == -1 {
		maxStr = "+inf"
	}

	minStr := fmt.Sprintf("%d", min)
	if min == -1 {
		minStr = "-inf"
	}

	vals, err = redis.Strings(conn.Do("ZREVRANGEBYSCORE", rkey, maxStr, minStr, "withscores", "limit", offset, count))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "range error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZCount(key string, min, max int64) (cnt int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	maxStr := fmt.Sprintf("%d", max)
	if max == -1 {
		maxStr = "+inf"
	}

	minStr := fmt.Sprintf("%d", min)
	if min == -1 {
		minStr = "-inf"
	}

	cnt, err = redis.Int(conn.Do("ZCOUNT", rkey, minStr, maxStr))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zcount error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZAdd(key string, score int64, member string) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	_, err = redis.Int(conn.Do("ZADD", rkey, score, member))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zadd error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZMulAdd(key string, mem2score map[int64]string) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	params := make([]interface{}, 0)
	params = append(params, rkey)
	for k, v := range mem2score {
		params = append(params, fmt.Sprintf("%d", k))
		params = append(params, fmt.Sprintf("%s", v))
	}

	_, err = redis.Int(conn.Do("ZADD", params...))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZMSET, "zmuladd error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZScore(key string, member string) (score int64, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	score = 0

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	val, err := redis.String(conn.Do("zscore", rkey, member))
	if err != nil {

		if err == redis.ErrNil {

			err = rest.NewAPIError(constant.E_REDIS_KEY_NO_EXIST, "key not exist,"+err.Error())
			return

		} else {
			err = rest.NewAPIError(constant.E_REDIS_GET, "get error,"+err.Error())
			log.Errorf(err.Error())
			return
		}
	}

	score, err = strconv.ParseInt(val, 10, 64)

	return
}

func (this *RedisApp) ZIncrBy(key string, member string, incr int64) (score int64, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	score = 0

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	val, err := redis.String(conn.Do("zincrby", rkey, incr, member))
	if err != nil {

		if err == redis.ErrNil {
			err = rest.NewAPIError(constant.E_REDIS_KEY_NO_EXIST, "key not exist,"+err.Error())
			return
		} else {
			err = rest.NewAPIError(constant.E_REDIS_GET, "get error,"+err.Error())
			log.Errorf(err.Error())
			return
		}
	}

	score, err = strconv.ParseInt(val, 10, 64)

	return
}

func (this *RedisApp) ZAddMul(key string, members map[string]int64) (err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	params := make([]interface{}, 0)
	params = append(params, rkey)

	for member, score := range members {
		params = append(params, fmt.Sprintf("%d", score))
		params = append(params, member)
	}

	_, err = redis.Int(conn.Do("ZADD", params...))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zaddmul error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZCard(key string) (total int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	total, err = redis.Int(conn.Do("ZCARD", rkey))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zcard error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRem(key string, member string) (remCnt int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	remCnt, err = redis.Int(conn.Do("ZREM", rkey, member))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zrem error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRemRangeByRank(key string, start, stop int) (remCnt int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	remCnt, err = redis.Int(conn.Do("ZREMRANGEBYRANK", rkey, start, stop))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zremrangebyrank error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) ZRemRangeByScore(key string, min, max int64) (remCnt int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	remCnt, err = redis.Int(conn.Do("ZREMRANGEBYSCORE", rkey, min, max))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_ZSET, "zremrangebyrank error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) HMGet(key string, fields []string) (vals map[string]string, err error) {

	if len(fields) == 0 {
		return
	}

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	strs := make([]interface{}, 0)
	strs = append(strs, rkey)
	for _, field := range fields {
		strs = append(strs, field)
	}

	replys, err := redis.Values(conn.Do("hmget", strs...))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_HMGET, "hmget error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	scan_dsts := make([]string, len(fields))

	err = redis.ScanSlice(replys, &scan_dsts)
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_HMGET, "hmget scan error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	vals = make(map[string]string)

	for i, reply := range replys {
		if reply == nil {
			continue
		}

		vals[fields[i]] = scan_dsts[i]
	}

	return
}

func (this *RedisApp) HMSet(key string, vals map[string]string) (err error) {

	if len(vals) == 0 {
		return
	}

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	strs := make([]interface{}, 0)
	strs = append(strs, rkey)
	for k, v := range vals {
		strs = append(strs, k)
		strs = append(strs, v)
	}

	reply, err := redis.String(conn.Do("hmset", strs...))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_HMSET, "hmset error "+err.Error())
		return
	}

	if reply != "OK" {
		err = rest.NewAPIError(constant.E_REDIS_HMSET, "hmset error "+reply)
		return
	}

	return
}

func (this *RedisApp) HGetAll(key string) (vals map[string]string, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	vals, err = redis.StringMap(conn.Do("hgetall", rkey))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_HGETALL, "hgetall error,"+err.Error())
		log.Errorf(err.Error())
		return
	}

	return
}

func (this *RedisApp) HIncrBy(key string, field string, cnt int) (val int, err error) {

	conn, err := this.GetConn()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer conn.Close()

	rkey := fmt.Sprintf("%s_%s", this.AppId, key)

	reply, err := redis.Int(conn.Do("hincrby", rkey, field, cnt))
	if err != nil {
		err = rest.NewAPIError(constant.E_REDIS_INCR, "incr error "+err.Error())
		return
	}

	val = reply

	return
}
