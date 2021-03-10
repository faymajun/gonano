package udp

import (
	"sync"

	"golang.org/x/net/ipv4"

	"github.com/pkg/errors"
)

const (
	batchSize = 16
)

var (
	// a system-wide packet buffer shared among sending, receiving and FEC
	// to mitigate high-frequency memory allocation for packets
	xmitBuf sync.Pool
)

func init() {
	xmitBuf.New = func() interface{} {
		return make([]byte, mtuLimit)
	}
}

type batchConn interface {
	ReadBatch(ms []ipv4.Message, flags int) (int, error)
	WriteBatch(ms []ipv4.Message, flags int) (int, error)
}

func (s *UDPSession) defaultReadLoop() {
	buf := make([]byte, mtuLimit)
	var src string
	for {
		if n, addr, err := s.conn.ReadFrom(buf); err == nil {
			// make sure the packet is from the same source
			if src == "" { // set source address
				src = addr.String()
			} else if addr.String() != src {
				continue
			}

			if n >= s.headerSize {
				s.packetInput(buf[:n])
			}
		} else {
			s.notifyReadError(errors.WithStack(err))
			return
		}
	}
}

func (s *UdpServer) defaultMonitor() {
	buf := make([]byte, mtuLimit)
	for {
		if n, from, err := s.conn.ReadFrom(buf); err == nil {
			if n >= s.headerSize {
				s.packetInput(buf[:n], from)
			}
		} else {
			s.notifyReadError(errors.WithStack(err))
			return
		}
	}
}

func (s *UDPSession) defaultTx(txqueue []ipv4.Message) {
	nbytes := 0
	npkts := 0
	for k := range txqueue {
		if n, err := s.conn.WriteTo(txqueue[k].Buffers[0], txqueue[k].Addr); err == nil {
			nbytes += n
			npkts++
			xmitBuf.Put(txqueue[k].Buffers[0])
		} else {
			s.notifyWriteError(errors.WithStack(err))
			break
		}
	}
}
