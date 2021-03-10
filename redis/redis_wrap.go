package redis

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/core/scheduler"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var (
	RedisMgr *RedisManager = &RedisManager{
		subMap:      map[string]*Redis{},
		scriptMap:   map[int]string{},
		scriptIndex: 0,
		stop:        0,
	}

	logger = logrus.WithField("component", "redis")
)

type (
	RedisManager struct {
		dbs         sync.Map
		subMap      map[string]*Redis
		scriptMap   map[int]string
		scriptIndex uint32
		stop        int32
	}

	Redis struct {
		*redis.Client
		pubsub        *redis.PubSub
		conf          *RedisConfig
		scriptHashMap map[int]string
		channels      []string
		fun           func(channel, data string)
	}

	RedisConfig struct {
		Addr     string
		Passwd   string
		DB       int
		PoolSize int
		Name     string
	}
)

func (rs *Redis) ScriptStr(cmd int, keys []string, args ...interface{}) (string, error) {
	data, err := rs.Script(cmd, keys, args...)
	if err != nil {
		return "", err
	}
	_, ok := data.(int64)
	if ok {
		return "", errors.New(fmt.Sprintf("redis script do err:%s", data))
	}
	str, ok := data.(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("redis script ret err:%s", data))
	}
	return str, nil
}

func (rs *Redis) ScriptStrArray(cmd int, keys []string, args ...interface{}) ([]string, error) {
	data, err := rs.Script(cmd, keys, args...)
	if err != nil {
		return nil, err
	}
	_, ok := data.(int64)
	if ok {
		return nil, errors.New(fmt.Sprintf("redis script array do err:%s", data))
	}
	iArray, ok := data.([]interface{})
	if !ok {
		return nil, errors.New(fmt.Sprintf("redis script array rets err:%s", data))
	}
	strArray := []string{}
	for _, v := range iArray {
		if str, ok := v.(string); ok {
			strArray = append(strArray, str)
		} else {
			return nil, errors.New(fmt.Sprintf("redis script array ret err:%s", data))
		}
	}
	return strArray, nil
}

func (rs *Redis) ScriptInt64(cmd int, keys []string, args ...interface{}) (int64, error) {
	data, err := rs.Script(cmd, keys, args...)
	if err != nil {
		return 0, err
	}
	code, ok := data.(int64)
	if ok {
		return code, nil
	}
	return 0, errors.New(fmt.Sprintf("redis script in64 ret err:%s", data))
}

func (rs *Redis) Script(cmd int, keys []string, args ...interface{}) (interface{}, error) {
	hash, _ := rs.scriptHashMap[cmd]
	re, err := rs.EvalSha(hash, keys, args...).Result()
	if err != nil {
		script, ok := RedisMgr.scriptMap[cmd]
		if !ok {
			return nil, errors.New(fmt.Sprintf("redis script error cmd not found cmd:%v", cmd))
		}
		if !strings.HasPrefix(err.Error(), "NOSCRIPT ") {
			return nil, err
		}
		hash, err = rs.ScriptLoad(script).Result()
		if err != nil {
			return nil, err
		}
		rs.scriptHashMap[cmd] = hash
		re, err = rs.EvalSha(hash, keys, args...).Result()
		if err == nil {
			return re, nil
		}
		return nil, err
	}
	return re, nil
}

func (rs *Redis) loadScript() {
	for cmd, script := range RedisMgr.scriptMap {
		hash, err := rs.ScriptLoad(script).Result()
		if err != nil {
			logger.Errorf("redis script load cmd:%v errstr:%v", RedisMgr.scriptMap[cmd], err)
			continue
		}
		rs.scriptHashMap[cmd] = hash
	}
}

func (rs *Redis) Sub(fun func(channel, data string), channels ...string) {
	rs.channels = channels
	rs.fun = fun
	if rs.pubsub != nil {
		rs.pubsub.Close()
	}

	rs.pubsub = rs.Subscribe(channels...)
	routine.Go(func() {
		for RedisMgr.IsRunning() {
			msg, err := rs.pubsub.ReceiveMessage()
			if err == nil {
				scheduler.PushTask(func() { fun(msg.Channel, msg.Payload) })
			} else if _, ok := err.(net.Error); !ok {
				break
			}
		}
	})

}

func (mgr *RedisManager) GetRedis(name string) *Redis {
	db, ok := mgr.dbs.Load(name)
	if !ok {
		logger.Errorf("redis get failed:%s", name)
		return nil
	}
	rds, ok := db.(*Redis)
	if !ok {
		logger.Errorf("redis get type failed:%s", name)
		return nil
	}
	return rds
}

func (mgr *RedisManager) Exist(name string) bool {
	_, ok := mgr.dbs.Load(name)
	return ok
}

func (mgr *RedisManager) Add(conf *RedisConfig) {
	if mgr.Exist(conf.Name) {
		logger.Errorf("redis already have rname:%s %s", conf.Name, conf.Addr)
		return
	}

	rds := &Redis{
		Client: redis.NewClient(&redis.Options{
			Addr:        conf.Addr,
			Password:    conf.Passwd,
			PoolSize:    conf.PoolSize,
			DB:          conf.DB,
			ReadTimeout: time.Second * 300,
		}),
		conf:          conf,
		scriptHashMap: map[int]string{},
	}
	rds.loadScript()

	if _, ok := mgr.subMap[conf.Addr]; !ok {
		mgr.subMap[conf.Addr] = rds
	}
	mgr.dbs.Store(conf.Name, rds)
	logger.Infof("connect to redis:%s %s", conf.Name, conf.Addr)
}

func (mgr *RedisManager) Del(name string) bool {
	rds := mgr.GetRedis(name)
	if rds == nil {
		return false
	}
	mgr.dbs.Delete(name)
	rds.Close()
	logger.Errorf("del redis from mgr:%s", name)
	return true
}

func (mgr *RedisManager) Stop() {
	atomic.AddInt32(&mgr.stop, 1)
	mgr.dbs.Range(func(k, r interface{}) bool {
		rds, ok := r.(*Redis)
		if !ok {
			logger.Errorf("redis close type error:%v", k)
		}
		if rds.pubsub != nil {
			rds.pubsub.Close()
		}
		rds.Close()
		return true
	})
}

func (mgr *RedisManager) IsRunning() bool {
	return atomic.LoadInt32(&mgr.stop) == 0
}

func (mgr *RedisManager) NewRedisScript(commit, str string) int {
	cmd := int(atomic.AddUint32(&mgr.scriptIndex, 1))
	mgr.scriptMap[cmd] = str
	return cmd
}

func StopRedis() {
	logger.Infof("Stop redis mgr!")
	RedisMgr.Stop()
}
