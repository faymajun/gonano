package util

import (
	"github.com/sirupsen/logrus"
	"net"
	"os"
)

var logger = logrus.WithField("component", "config")

// LocalIP 获取本地ip
func LocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, nil
}

// LocalIPString 获取本地ip string
func LocalIPString() string {
	ip, err := LocalIP()
	if err != nil {
		logger.Warn("Error determining local ip address. ", err)
		return ""
	}
	if ip == nil {
		logger.Warn("Could not determine local ip address")
		return ""
	}
	return ip.String()
}

func HostName() string {
	if host, err := os.Hostname(); err != nil {
		return LocalIPString()
	} else {
		return host
	}
}
