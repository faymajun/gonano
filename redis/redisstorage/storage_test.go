package redisstorage

import (
	"fmt"
	"testing"

	"github.com/faymajun/gonano/redis"
)

func TestInvalidRoleId(t *testing.T) {
	redisConf := &redis.RedisConfig{Addr: "192.168.9.196:6670", Passwd: "", PoolSize: 10, Name: Global_Redis}
	redis.InitRedisManager(redisConf)
	fmt.Println(InvalidRoleId(10000))
	fmt.Println(InvalidRoleId(1000100001))
	fmt.Println(InvalidRoleId(1010000001))
}
