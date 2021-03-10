package redis

import (
	"github.com/faymajun/gonano/core"
)

const (
	Local_Redis = "LocalRedis"
)

func InitDefaultConfig(addr, passwd string, db int) {
	redisConf := &RedisConfig{Addr: addr, Passwd: passwd, DB: db, PoolSize: 0, Name: Local_Redis}
	InitRedisManager(redisConf)
}

func InitRedisManager(config *RedisConfig) {
	RedisMgr.Add(config)
}

func LocalRedis() *Redis {
	return RedisMgr.GetRedis(Local_Redis)
}

func GetRedis(name string) *Redis {
	return RedisMgr.GetRedis(name)
}

func Add(cfg *RedisConfig) {
	RedisMgr.Add(cfg)
}

func Del(name string) {
	RedisMgr.Del(name)
}

func Exist(name string) bool {
	return RedisMgr.Exist(name)
}

func NewRedisScript(commit, str string) int {
	return RedisMgr.NewRedisScript(commit, str)
}

func GetGrowthChannel(id int64) string {
	return core.Sprintf("Growth_Channel_%d", id)
}
