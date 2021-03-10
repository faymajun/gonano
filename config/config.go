package config

import (
	"time"

	"github.com/spf13/viper"
)

var Content content

type content struct{}

func (content) String(key string) string {
	return viper.GetString(key)
}

func (content) Int(key string) int {
	return viper.GetInt(key)
}

func (content) Int64(key string) int64 {
	return viper.GetInt64(key)
}

func (content) Bool(key string) bool {
	return viper.GetBool(key)
}

var ServerConfig serverConfig

type serverConfig struct {
	secret      []byte
	id          int32
	closeReq    bool
	version     string
	maxLimit    int32
	heartbeat   int64
	serverID    string
	startupTime int64 // 服务器启动时间
	closing     bool
	roleMaxNum  int64
}

func Initialize() {
	ServerConfig.secret = []byte(Content.String("token_secret"))
	ServerConfig.id = int32(Content.Int("core_id"))
	ServerConfig.heartbeat = Content.Int64("core_heartbeat")
	ServerConfig.version = Content.String("core_version")
	ServerConfig.maxLimit = int32(Content.Int("login_maxlimit"))
	ServerConfig.closeReq = Content.Bool("login_closereq")
	ServerConfig.serverID = Content.String("core_serverID")
	ServerConfig.roleMaxNum = int64(Content.Int("maxRoleNum"))

	if ServerConfig.heartbeat <= 0 { // 默认客户端心跳值
		ServerConfig.heartbeat = 60
	}

	ServerConfig.startupTime = time.Now().Unix()
}

func Secret() []byte {
	return ServerConfig.secret
}

func ServerId() int32 {
	return ServerConfig.id
}

func ServerID() string {
	return ServerConfig.serverID
}

func Heartbeat() int64 {
	return ServerConfig.heartbeat
}

func MaxLimit() int32 {
	return ServerConfig.maxLimit
}

func RoleMaxNum() int64 {
	return ServerConfig.roleMaxNum
}

func Version() string {
	return ServerConfig.version
}

func CloseReq() bool {
	return ServerConfig.closeReq
}

//设置是否能创建角色
func SetCloseReq(isClose bool) {
	ServerConfig.closeReq = isClose
}

//设置最大登录人数
func SetLoginMaxLimit(num int32) {
	ServerConfig.maxLimit = num
}

//设置最大创建角色数
func SetRoleMaxNum(num int64) {
	ServerConfig.roleMaxNum = num
}

func SetServerID(serverId string) {
	ServerConfig.serverID = serverId
}

func StartUpTime() int64 {
	return ServerConfig.startupTime
}

func SetClosing() {
	ServerConfig.closing = true
}

func Closing() bool {
	return ServerConfig.closing
}
