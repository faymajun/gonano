package udp

import (
	"github.com/faymajun/gonano/core"
)

type clientHandler struct {
}

func (client *clientHandler) OnSessionCreate(session *core.Session) bool {
	logger.Infof("Session Create")
	return true
}

func (client *clientHandler) OnSessionClose(session *core.Session) {
	logger.Infof("Session Close")
	return
}
