package logx

import (
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

func ErrorStack(msg string) {
	logrus.Errorf("%s\n%s", msg, debug.Stack())
}
