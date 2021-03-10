package udp

import (
	"fmt"
	"github.com/faymajun/gonano/core/net/session"
	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/core/scheduler"
	"github.com/faymajun/gonano/message"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	. "github.com/faymajun/gonano/core/net"
)

type (
	UDPSession struct {
		conn  net.PacketConn
		s     *UdpServer
		coder Codec

		handler     SessionHandler
		userSession *session.Session

		read  chan *packet.RecvMessage
		write chan *packet.SendMessage

		remote     net.Addr
		headerSize int

		// notifications
		dieOnce int32

		// socket error handling
		socketReadError      atomic.Value
		socketWriteError     atomic.Value
		chSocketReadError    chan struct{}
		chSocketWriteError   chan struct{}
		socketReadErrorOnce  sync.Once
		socketWriteErrorOnce sync.Once

		xconn           batchConn // for x/net
		xconnWriteError error

		mu sync.Mutex
	}

	setReadBuffer interface {
		SetReadBuffer(bytes int) error
	}

	setWriteBuffer interface {
		SetWriteBuffer(bytes int) error
	}
)

func DialWithOptions(raddr string, waddr string, coder Codec, handler SessionHandler) (*UDPSession, error) {
	// network type detection
	udpaddr, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	network := "udp4"
	if udpaddr.IP.To4() == nil {
		network = "udp"
	}

	conn, err := net.ListenUDP(network, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewConn(waddr, conn, coder, handler)
}

func NewConn(waddr string, conn net.PacketConn, coder Codec, handler SessionHandler) (*UDPSession, error) {
	udpaddr, err := net.ResolveUDPAddr("udp", waddr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sess := newUDPSession(nil, conn, udpaddr, coder, handler)
	sess.handler.OnSessionCreate(sess.userSession)
	return sess, nil
}

// newUDPSession create a new udp session for client or server
func newUDPSession(s *UdpServer, conn net.PacketConn, remote net.Addr, coder Codec, handler SessionHandler) *UDPSession {
	sess := new(UDPSession)
	sess.chSocketReadError = make(chan struct{})
	sess.chSocketWriteError = make(chan struct{})
	sess.remote = remote
	sess.conn = conn
	sess.s = s
	sess.coder = coder
	sess.handler = handler
	sess.userSession = session.NewSession(sess)
	sess.read = make(chan *packet.RecvMessage, WriteQueLen) // udp发送较多设置64个
	sess.write = make(chan *packet.SendMessage, WriteQueLen)
	sess.dieOnce = 0

	// cast to writebatch conn
	addr, _ := net.ResolveUDPAddr("udp", conn.LocalAddr().String())
	if addr.IP.To4() != nil {
		sess.xconn = ipv4.NewPacketConn(conn)
	} else {
		sess.xconn = ipv6.NewPacketConn(conn)
	}

	if sess.s == nil { // it's a client connection
		routine.Go(func() {
			sess.readLoop()
		})
	}

	return sess
}

func (s *UDPSession) SessionID() uint32 {
	return 0
}

func (s *UDPSession) UserSession() *session.Session {
	return s.userSession
}

func (s *UDPSession) Write() error {
	for {
		select {
		case msg, ok := <-s.write:
			if !ok {
				return fmt.Errorf("write is closed")
			}

			var data []byte
			if msg.DontEncode == false {
				payload, err := msg.Serialize()
				if err != nil {
					return fmt.Errorf("serialize msg:%d failed:%s", msg.MsgID, err)
				}
				data = s.coder.Encode(msg.MsgID, payload)
			} else {
				data = msg.Payload.([]byte)
			}

			_, err := s.conn.WriteTo(data, s.remote)
			if err != nil {
				s.notifyWriteError(errors.WithStack(err))
			}
		}
	}
}

func (s *UDPSession) Read(heartbeat time.Duration) error {
	// deadline for current reading operation
	var timeout *time.Timer
	var c <-chan time.Time

	for {
		// wait for read event or timeout or error
		select {
		case pack, ok := <-s.read:
			if !ok {
				return errors.WithStack(errClose)
			}
			pack.Session = s.userSession
			scheduler.PushPacket(pack)

			if timeout != nil {
				timeout.Stop()
			}

			if heartbeat > 0 {
				timeout = time.NewTimer(heartbeat)
				c = timeout.C
			}
		case <-c:
			return errors.WithStack(errTimeout)
		case <-s.chSocketReadError:
			return s.socketReadError.Load().(error)
		}
	}
}

// Close closes the connection.
func (s *UDPSession) Close() error {
	if !atomic.CompareAndSwapInt32(&s.dieOnce, 0, 1) {
		return errors.WithStack(ErrDupClosed)
	}

	close(s.read)
	close(s.write)

	scheduler.PushTask(func() { s.handler.OnSessionClose(s.userSession) })

	if s.s != nil { // belongs to server
		s.s.closeSession(s.remote)
		return nil
	} else { // client socket close
		return s.conn.Close()
	}
}

func (s *UDPSession) LocalAddr() net.Addr { return s.conn.LocalAddr() }

func (s *UDPSession) RemoteAddr() net.Addr { return s.remote }

// SetDSCP sets the 6bit DSCP field in IPv4 header, or 8bit Traffic Class in IPv6 header.
//
// It has no effect if it's accepted from UdpServer.
func (s *UDPSession) SetDSCP(dscp int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		if nc, ok := s.conn.(net.Conn); ok {
			addr, _ := net.ResolveUDPAddr("udp", nc.LocalAddr().String())
			if addr.IP.To4() != nil {
				return ipv4.NewConn(nc).SetTOS(dscp << 2)
			} else {
				return ipv6.NewConn(nc).SetTrafficClass(dscp)
			}
		}
	}
	return errInvalidOperation
}

// SetReadBuffer sets the socket read buffer, no effect if it's accepted from UdpServer
func (s *UDPSession) SetReadBuffer(bytes int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		if nc, ok := s.conn.(setReadBuffer); ok {
			return nc.SetReadBuffer(bytes)
		}
	}
	return errInvalidOperation
}

// SetWriteBuffer sets the socket write buffer, no effect if it's accepted from UdpServer
func (s *UDPSession) SetWriteBuffer(bytes int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		if nc, ok := s.conn.(setWriteBuffer); ok {
			return nc.SetWriteBuffer(bytes)
		}
	}
	return errInvalidOperation
}

func (s *UDPSession) notifyReadError(err error) {
	s.socketReadErrorOnce.Do(func() {
		s.socketReadError.Store(err)
		close(s.chSocketReadError)
	})
}

func (s *UDPSession) notifyWriteError(err error) {
	s.socketWriteErrorOnce.Do(func() {
		s.socketWriteError.Store(err)
		close(s.chSocketWriteError)
	})
}

// packet input stage
func (s *UDPSession) packetInput(data []byte) {
	if atomic.LoadInt32(&s.dieOnce) != 0 {
		return
	}

	if len(data) == 0 {
		return
	}

	// 防止udp接受主线程阻塞
	if len(s.read) >= cap(s.read) {
		logger.Warn("read buffer exceed, session close", s.RemoteAddr())
		return
	}

	// read buf push main process
	pack, err := s.coder.Decode(data)
	if err != nil {
		logger.Errorf("decode data err:%s, addr=%s", err, s.RemoteAddr())
	}

	if pack != packet.EmptyRecvPack {
		routine.Try(func() { s.read <- &pack }, nil)
	}
}

func (s *UDPSession) Send(msgid message.MSGID, pbmsg proto.Message) error {
	return s.send(&packet.SendMessage{MsgID: msgid, Payload: pbmsg})
}

func (s *UDPSession) SendData(msgid message.MSGID, data []byte) error {
	return s.send(&packet.SendMessage{MsgID: msgid, Payload: data})
}

func (s *UDPSession) SendDataNotEncode(msgid message.MSGID, data []byte) error {
	return s.send(&packet.SendMessage{MsgID: msgid, Payload: data, DontEncode: true})
}

func (s *UDPSession) send(msg *packet.SendMessage) error {
	if atomic.LoadInt32(&s.dieOnce) != 0 {
		return nil
	}

	if len(s.write) >= cap(s.write) {
		logger.Warn("send buffer exceed, session close", s.RemoteAddr())

		// 这里不用断开，有tcp控制，就不发就可以了
		//s.Close()
		return fmt.Errorf("send buffer excced: %s", s.RemoteAddr())
	}

	//if tags.DEBUG {
	//	if msg.MsgID != message.MSGID_ResHeartbeatE &&
	//		msg.MsgID != message.MSGID_ReqHeartbeatE &&
	//		msg.MsgID != message.MSGID_FrameMsgE &&
	//		msg.MsgID != message.MSGID_Growth_ReqHeartbeatE &&
	//		msg.MsgID != message.MSGID_Growth_ResHeartbeatE &&
	//		msg.MsgID != message.MSGID_ResUdpHeartBeatE {
	//		println(fmt.Sprintf(colorized.Magenta("==> %s, %+v"), msg.MsgID.String(), msg.Payload))
	//	}
	//}

	routine.Try(func() { s.write <- msg }, nil)
	return nil

}
