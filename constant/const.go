package constant

import (
	"github.com/faymajun/gonano/config"
	"github.com/faymajun/gonano/core"
)

const DefaultSplit = "_"
const GrowthNum = 100000
const ClubNum = 100000
const GrowthNumLen = 5

const (
	MinRoleId = 10000 * GrowthNum
	MinClubId = 10000 * ClubNum
)

func GetGrowthIdOnRoleId(roleId int64) int32 {
	return int32(roleId % GrowthNum)
}

func IsRobot(roleId int64) bool {
	return roleId < MinRoleId
}

func IsPlayerId(roleId int64) bool {
	return roleId > MinRoleId
}

func GetGrowthIDByRoleId(roleId int64) string {
	if roleId < MinRoleId {
		return ""
	}

	str := core.Sprintf("%d", roleId)
	return str[len(str)-5:]
}

func IsMine(roleId int64) bool {
	if GetGrowthIdOnRoleId(roleId) == config.ServerId() {
		return true
	}
	return false
}

func LoginAccount(accountId string) string {
	return core.Sprintf("login:Account:%s", accountId)
}

func GetClubServerIdOnClubId(clubId int64) int32 {
	return int32(clubId % ClubNum)
}

func GetClubServerIDByClubId(clubId int64) string {
	if clubId < MinClubId {
		return ""
	}

	str := core.Sprintf("%d", clubId)
	return str[len(str)-5:]
}

func IsMineClub(clubId int64) bool {
	if GetClubServerIdOnClubId(clubId) == config.ServerId() {
		return true
	}
	return false
}
