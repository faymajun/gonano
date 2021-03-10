package scheduler

import "github.com/faymajun/gonano/core/packet"

func Start() {
	scheduler.Start()
}

func Stop() {
	scheduler.Stop()
}

func PushTask(fn func()) {
	scheduler.PushTask(fn)
}

func PushPacket(pack *packet.RecvMessage) {
	scheduler.PushPacket(pack)
}
