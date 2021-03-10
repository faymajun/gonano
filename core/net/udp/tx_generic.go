// +build !linux

package udp

import "golang.org/x/net/ipv4"

func (s *UDPSession) readLoop() {
	s.defaultReadLoop()
}

func (s *UdpServer) monitor() {
	s.defaultMonitor()
}

func (s *UDPSession) tx(txqueue []ipv4.Message) {
	s.defaultTx(txqueue)
}
