package consul

import (
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var log = logrus.WithField("com", "consul")
var MySelf = &Cluster{}

type Cluster struct {
	Dc      string
	Cluster string
	Service string
	Index   int
}

type CallType int

// 这是调用配置相关结构
type CallOption struct {
	Timeout time.Duration
}

func makeServiceId(cluster string, service string, index int) string {
	return cluster + "-" + service + "-" + strconv.Itoa(index)
}

//func makeServiceIdWithDc(dc string, cluster string, service string, index int) string {
//	return dc + "-" + cluster + "-" + service + "-" + strconv.Itoa(index)
//}
