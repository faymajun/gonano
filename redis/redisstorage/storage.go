package redisstorage

import (
	"strconv"

	"github.com/faymajun/gonano/constant"
	"github.com/faymajun/gonano/core"

	goredis "github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

const (
	Idgen = "idgen"
)

func AccountRoleId(accountId string) string {
	return core.Sprintf("account:roleid:%s", accountId)
}

// 必须连接上global 数据库
func InvalidRoleId(roleId int64) bool {
	if constant.IsPlayerId(roleId) == false {
		return false
	}

	if GlobalRedis() == nil {
		return false
	}

	serverId := constant.GetGrowthIDByRoleId(roleId)
	field := core.Sprintf("roleId_%s", serverId)
	roleIdResult := GlobalRedis().HGet("idgen", field)
	err := roleIdResult.Err()
	if err != nil && err != goredis.Nil {
		logrus.Errorf("InvalidRoleId, Error=%v", err)
		return false
	}

	maxRoleId, err := strconv.ParseInt(roleIdResult.Val(), 10, 64)
	if err != nil || maxRoleId == 0 {
		logrus.Errorf("InvalidRoleId, ParseInt Error=%v", err)
		return false
	}

	return roleId/constant.GrowthNum <= maxRoleId
}
