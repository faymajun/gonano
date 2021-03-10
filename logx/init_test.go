package logx

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestInit(t *testing.T) {
	Init()
	logger.Trace("this is a test trace logx")
	logger.Debug("this is a test info logx")
	logger.Info("this is a test info logx")
	logger.Warn("this is a test warn logx")
	logger.Error("this is a test error logx")
	panic("test panic")
}

var logger = logrus.WithField("com", "logx")

func TestLogFiled(t *testing.T) {
	Init()
	logger.Trace("this is a test trace logx")
	logger.Debug("this is a test info logx")
	logger.Info("this is a test info logx")
	logger.Warn("this is a test warn logx")
	logger.Error("this is a test error logx")
}
