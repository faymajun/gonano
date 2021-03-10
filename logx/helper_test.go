package logx

import (
	"testing"
)

func TestErrorStack(t *testing.T) {
	Init()
	ErrorStack("123")
}
