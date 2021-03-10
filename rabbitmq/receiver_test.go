package rabbitmq

import (
	"github.com/faymajun/gonano/core/routine"
	"testing"
)

var addr = "amqp://root:lmd2019@192.168.8.216:5672/"
var name = "Chat"

func TestConsumer(t *testing.T) {
	routine.Go(func() {
		InitConsumer(addr, name)
	})
	routine.Go(func() {
		InitConsumer(addr, name)
	})
	InitConsumer(addr, name)
}
