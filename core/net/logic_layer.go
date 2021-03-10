package net

import (
	"github.com/faymajun/gonano/core/net/session"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("component", "net")

type SessionHandler interface {
	OnSessionCreate(session *session.Session) bool
	OnSessionClose(session *session.Session)
}
