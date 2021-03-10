package rabbitmq

import (
	"testing"
	"time"

	"github.com/faymajun/gonano/message"
)

func Test_PublishMsg(t *testing.T) {
	InitProducer(addr)
	AddQueues(name)
	PublishMsg(name, message.MSGID_ReqHeartbeatE, &message.ReqHeartbeat{time.Now().Unix()})
	StopProducer()
}
