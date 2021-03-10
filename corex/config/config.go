package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type TlsInfo struct {
	CaFile   string `yaml:"ca_file"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type EtcdInfo struct {
	Endpoints string     `yaml:"endpoints"`
	SslEnable bool       `yaml:"ssl_enable"`
	SslKey    SslKeyInfo `yaml:"ssl_key"`
}

type SslKeyInfo struct {
	CaFile      string `yaml:"ca_file"`
	SrvCertFile string `yaml:"srv_cert_file"`
	SrvKeyFile  string `yaml:"srv_key_file"`
	CltCertFile string `yaml:"clt_cert_file"`
	CltKeyFile  string `yaml:"clt_key_file"`
}

type ANetInfo struct {
	Network       string `yaml:"network"`
	Addr          string `yaml:"addr"`
	AdvertiseAddr string `yaml:"advertise_addr"`
	MaxEvents     int32  `yaml:"max_events"`
}

type HttpInfo struct {
	Mode string `yaml:"mode"`
	Host string `yaml:"host"`
	Port int32  `yaml:"port"`
	Key  string `yaml:"key"`
}

type AMQInfo struct {
	Host       string `yaml:"host"`
	Port       int32  `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Durable    bool   `yaml:"durable"`
	AutoDelete bool   `yaml:"auto_delete"`
	Exclusive  bool   `yaml:"exclusive"`
	NoWait     bool   `yaml:"no_wait"`
}

type MongoDB struct {
	Hosts      []string `yaml:"hosts"`
	DbName     string   `yaml:"dbame"`
	AuthSource string   `yaml:"authSource"`
	ReplicaSet string   `yaml:"replicaSet"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
}

type BattleParam struct {
	MatchTimeout         int32 `yaml:"match_timeout"`          // 匹配超时时间, by Milliseconds
	MatchReadyCountDown  int32 `yaml:"match_ready_countdown"`  // 匹配完成准备倒计时, by Milliseconds
	BattleReadyCountDown int32 `yaml:"battle_ready_countdown"` // 战斗开始前准备倒计时, by Milliseconds
	KeepAlive            int32 `yaml:"keepalive"`              // 心跳时间, by Milliseconds
}

type ServerInfo struct {
	Id       int32       `yaml:"id"`       // 服务id, 相同类型服务需要配置成不同的id
	Priority int32       `yaml:"priority"` // 战斗服分配优先级, 负载均衡用
	ANet     ANetInfo    `yaml:"anet"`     // 网络库配置
	Http     HttpInfo    `yaml:"http"`     // http服务配置
	MongoDB  MongoDB     `yaml:"mongodb"`  // 数据库连接配置
	Etcd     EtcdInfo    `yaml:"etcd"`     // etcd连接配置
	AMQ      AMQInfo     `yaml:"amq"`      // rabbitmq连接配置
	Params   BattleParam `yaml:"params"`   // 匹配/战斗服参数
}

//////////////
var (
	PWD  string
	_cfg ServerInfo
)

func init() {
	PWD, _ = os.Getwd()
}

func Load(cfgfile string) error {
	data, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal([]byte(data), &_cfg); err != nil {
		return err
	}
	return nil
}

func Config() ServerInfo {
	return _cfg
}
