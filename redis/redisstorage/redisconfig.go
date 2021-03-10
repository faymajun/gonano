package redisstorage

import (
	"fmt"
	"github.com/faymajun/gonano/config"
	"github.com/faymajun/gonano/core"
	"github.com/faymajun/gonano/core/coroutine"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/redis"
	"strings"

	go_redis "github.com/go-redis/redis"
)

const (
	Login_Redis      = "LoginRedis"
	Log_Redis        = "LogRedis"
	Global_Redis     = "GlobalRedis"
	Backs_Redis      = "BacksRedis"
	Club_Redis       = "ClubRedis"
	Club_Redis_Async = "ClubRedisAsync"
)

func InitLoginConfig() {
	redisConf := &redis.RedisConfig{Addr: config.Content.String("loginRedis_addr"),
		Passwd: config.Content.String("loginRedis_passwd"), DB: config.Content.Int("loginRedis_db"), PoolSize: 100, Name: Login_Redis}
	redis.InitRedisManager(redisConf)
}

func LoginRedis() *redis.Redis {
	return redis.RedisMgr.GetRedis(Login_Redis)
}

func InitGlobalConfig() {
	poolSize := config.Content.Int("globalRedis_poolSize")
	if poolSize < 0 {
		poolSize = 0
	}
	redisConf := &redis.RedisConfig{Addr: config.Content.String("globalRedis_addr"),
		Passwd: config.Content.String("globalRedis_passwd"), DB: config.Content.Int("globalRedis_db"), PoolSize: poolSize, Name: Global_Redis}
	redis.InitRedisManager(redisConf)
}

func GlobalRedis() *redis.Redis {
	return redis.RedisMgr.GetRedis(Global_Redis)
}

func InitLogConfig() {
	redisConf := &redis.RedisConfig{Addr: config.Content.String("logRedis_addr"),
		Passwd: config.Content.String("logRedis_passwd"), DB: config.Content.Int("logRedis_db"), PoolSize: 50, Name: Log_Redis}
	redis.InitRedisManager(redisConf)
}

func LogRedis() *redis.Redis {
	return redis.RedisMgr.GetRedis(Log_Redis)
}

const RoutineCount = 10 * 32

var routines []*coroutine.Coroutine

func InitClubConfig() {
	routines = make([]*coroutine.Coroutine, RoutineCount)
	for i := 0; i < RoutineCount; i++ {
		routines[i] = coroutine.NewCoroutine(64, int64(i))
	}
	redisConf := &redis.RedisConfig{Addr: config.Content.String("clubRedis_addr"),
		Passwd: config.Content.String("clubRedis_passwd"), DB: config.Content.Int("clubRedis_db"), PoolSize: 1, Name: Club_Redis}
	redis.InitRedisManager(redisConf)
	redisConfAsync := &redis.RedisConfig{Addr: config.Content.String("clubRedis_addr"),
		Passwd: config.Content.String("clubRedis_passwd"), DB: config.Content.Int("clubRedis_db"), Name: Club_Redis_Async}
	redis.InitRedisManager(redisConfAsync)
	ClubAsyncRedis().WrapProcess(func(old func(go_redis.Cmder) error) func(go_redis.Cmder) error {
		return func(cmd go_redis.Cmder) error {
			if len(cmd.Args()) >= 2 {
				key := fmt.Sprint(cmd.Args()[1])
				var sum int
				for _, v := range strings.Split(key, ":") {
					id := core.Atoi(v)
					if id != 0 {
						sum = id / 100000
						break
					}
				}
				if sum == 0 {
					bits := []uint8(key)
					for _, bit := range bits {
						sum = sum + int(bit)
					}
				}
				index := sum % RoutineCount
				_ = routines[index].PushTask(func() {
					_ = old(cmd)
					if cmd.Err() != nil {
						fmt.Println("async redis ", key, " err ", cmd.Err())
					}
				}, false)

			} else {
				routine.Go(func() {
					_ = old(cmd)
					if cmd.Err() != nil {
						fmt.Println("async redis err ", cmd.Err())
					}
				})
			}
			return nil
		}
	})
}

func ClubRedis() *redis.Redis {
	return redis.RedisMgr.GetRedis(Club_Redis)
}

func ClubAsyncRedis() *redis.Redis {
	return redis.RedisMgr.GetRedis(Club_Redis_Async)
}

func InitBacksConfig() {
	redisConf := &redis.RedisConfig{Addr: config.Content.String("backsRedis_addr"),
		Passwd: config.Content.String("backsRedis_passwd"), DB: config.Content.Int("backsRedis_db"), PoolSize: 10, Name: Backs_Redis}
	redis.InitRedisManager(redisConf)
}

func BacksRedis() *redis.Redis {
	return redis.RedisMgr.GetRedis(Backs_Redis)
}
