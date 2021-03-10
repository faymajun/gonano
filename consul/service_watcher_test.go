package consul

import (
	"github.com/faymajun/gonano/logx"
	"testing"
)

func TestWatchServer(t *testing.T) {
	// 日志初始化
	logx.Init()
	log.Info("test start!!!")

	WatchServer("dc1", "dc1", "gateway")

	select {}

}
