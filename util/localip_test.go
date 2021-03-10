package util

import (
	"testing"
)

func TestLocalIP(t *testing.T) {
	log.Print(LocalIP())
}

func TestLocalIPString(t *testing.T) {
	log.Print(LocalIPString())
}

func TestHostName(t *testing.T) {
	log.Print(HostName())
}
