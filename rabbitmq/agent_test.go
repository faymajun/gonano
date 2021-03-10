package rabbitmq

import (
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/message"
	"testing"
	"time"
)

func TestInitAgent(t *testing.T) {
	routine.Go(func() {
		InitAgent(addr, name)
	})
	routine.Go(func() {
		InitAgent(addr, name)
	})
	InitAgent(addr, name)
}

func TestInitAgentPush(t *testing.T) {
	routine.Go(func() {
		InitAgent(addr, name)
	})
	time.Sleep(time.Second)
	Agent.PublishMsg(name, message.MSGID_ReqHeartbeatE, &message.ReqHeartbeat{time.Now().Unix()})
}
