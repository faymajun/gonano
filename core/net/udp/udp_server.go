package udp

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/core/routine"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/pkg/errors"

	. "github.com/faymajun/gonano/core/net"
)

const (
	// maximum packet size
	mtuLimit = 1500
)

var (
	errInvalidOperation = errors.New("invalid operation")
	errTimeout          = errors.New("timeout")
	logger              = logrus.WithField("net", "udp")
	errClose            = errors.New("closed")
)

type (
	UdpServer struct {
		conn net.PacketConn

		sessions    map[string]*UDPSession
		sessionLock sync.Mutex
		headerSize  int

		die     chan struct{} // notify the server has closed
		dieOnce sync.Once

		// socket error handling
		socketReadError     atomic.Value
		chSocketReadError   chan struct{}
		socketReadErrorOnce sync.Once

		coder   Codec
		handler SessionHandler

		heartbeat int64
	}
)

func StartWithOptions(saddr string, coder Codec, handler SessionHandler, heartbeat int64) (*UdpServer, error) {
	udpaddr, err := net.ResolveUDPAddr("udp", saddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ServeConn(conn, coder, handler, heartbeat)
}

func ServeConn(conn net.PacketConn, coder Codec, handler SessionHandler, heartbeat int64) (*UdpServer, error) {
	s := new(UdpServer)
	s.conn = conn
	s.coder = coder
	s.handler = handler
	s.sessions = make(map[string]*UDPSession)
	s.die = make(chan struct{})
	s.chSocketReadError = make(chan struct{})

	s.heartbeat = heartbeat

	routine.Go(func() {
		s.monitor()
	})

	return s, nil
}

func (s *UdpServer) notifyReadError(err error) {
	s.socketReadErrorOnce.Do(func() {
		s.socketReadError.Store(err)
		close(s.chSocketReadError)

		// propagate read error to all sessions
		s.sessionLock.Lock()
		for _, s := range s.sessions {
			s.notifyReadError(err)
		}
		s.sessionLock.Unlock()
	})
}

func (s *UdpServer) SetReadBuffer(bytes int) error {
	if nc, ok := s.conn.(setReadBuffer); ok {
		return nc.SetReadBuffer(bytes)
	}
	return errInvalidOperation
}

func (s *UdpServer) SetWriteBuffer(bytes int) error {
	if nc, ok := s.conn.(setWriteBuffer); ok {
		return nc.SetWriteBuffer(bytes)
	}
	return errInvalidOperation
}

// SetDSCP sets the 6bit DSCP field in IPv4 header, or 8bit Traffic Class in IPv6 header.
func (s *UdpServer) SetDSCP(dscp int) error {
	if nc, ok := s.conn.(net.Conn); ok {
		addr, _ := net.ResolveUDPAddr("udp", nc.LocalAddr().String())
		if addr.IP.To4() != nil {
			return ipv4.NewConn(nc).SetTOS(dscp << 2)
		} else {
			return ipv6.NewConn(nc).SetTrafficClass(dscp)
		}
	}
	return errInvalidOperation
}

func (s *UdpServer) Close() error {
	var once bool
	s.dieOnce.Do(func() {
		close(s.die)
		once = true
	})

	if once {
		return s.conn.Close()
	} else {
		return errors.WithStack(io.ErrClosedPipe)
	}
}

func (s *UdpServer) closeSession(remote net.Addr) (ret bool) {
	s.sessionLock.Lock()
	defer s.sessionLock.Unlock()
	if _, ok := s.sessions[remote.String()]; ok {
		delete(s.sessions, remote.String())
		return true
	}
	return false
}

func (s *UdpServer) Addr() net.Addr { return s.conn.LocalAddr() }

// monotonic reference time point
var refTime time.Time = time.Now()

// currentMs returns current elasped monotonic milliseconds since program startup
func currentMs() uint32 { return uint32(time.Now().Sub(refTime) / time.Millisecond) }

// packet input stage
func (s *UdpServer) packetInput(data []byte, addr net.Addr) {
	s.sessionLock.Lock()
	sess, ok := s.sessions[addr.String()]
	s.sessionLock.Unlock()

	if !ok { // new address:port
		sess = s.addClient(addr)
		sess.packetInput(data)

	} else {
		sess.packetInput(data)
	}
}

func (s *UdpServer) addClient(addr net.Addr) *UDPSession {
	sess := s.newClient(addr)
	s.sessionLock.Lock()
	s.sessions[addr.String()] = sess
	s.sessionLock.Unlock()
	return sess
}

func (s *UdpServer) newClient(addr net.Addr) *UDPSession {
	logger.Debugf("udp new session: %v", addr)
	sess := newUDPSession(s, s.conn, addr, s.coder, s.handler)

	routine.Go(func() {
		heartbeat := time.Duration(s.heartbeat) * time.Second
		err := sess.Read(heartbeat)
		logger.Warn(err)
		sess.Close()
	})

	routine.Go(func() {
		err := sess.Write()
		logger.Warn(err)
		sess.Close()
	})

	return sess
}
