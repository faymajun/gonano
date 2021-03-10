package anet

import (
	log "github.com/sirupsen/logrus"
	"net"
	"time"

	"github.com/xtaci/kcp-go"
)

const (
	ID_POOL_SIZE = 32
)

type Server struct {
	net      string
	addr     string
	listener net.Listener
	conn     net.PacketConn
	eq       chan Event
	idpool   chan int64
	proto    Protocol
}

func NewServer(net string, addr string, proto Protocol, eq chan Event) *Server {
	srv := Server{
		net:      net,
		addr:     addr,
		listener: nil,
		eq:       eq,
		idpool:   make(chan int64, ID_POOL_SIZE),
		proto:    proto,
	}
	return &srv
}

func (s *Server) ListenAndServe() error {
	if s.net == "kcp" {
		log.Info("kcp: ", s.addr)
		//listener, err := kcp.ListenWithOptions(s.addr, nil, 10, 3)
		listener, err := kcp.ListenWithOptions(s.addr, nil, 0, 0)
		if err != nil {
			return err
		}
		//listener.SetDSCP(0)
		if err := listener.SetReadBuffer(4194304); err != nil {
			return err
		}
		if err := listener.SetWriteBuffer(4194304); err != nil {
			return err
		}
		s.listener = listener
	} else if s.net == "tcp" {
		if tcpAddr, err := net.ResolveTCPAddr(s.net, s.addr); err != nil {
			return err
		} else {
			s.listener, err = net.ListenTCP(s.net, tcpAddr)
		}
	} else if s.net == "udp" {
		var err error
		s.conn, err = net.ListenPacket(s.net, s.addr)
		if err != nil {
			return err
		}
		log.Info("listen on: ", s.addr)

		session := newUDPSession(1001, s.conn, s.proto)
		s.eq <- newEvent(EVENT_ACCEPT, session, nil)
		session.Start(s.eq)
		return nil
	}
	go func() {
		id := int64(1)
		for {
			s.idpool <- id
			id++
		}
	}()
	go func() {
		defer s.listener.Close()
		var tempDelay time.Duration // how long to sleep on accept failure
		for {
			conn, e := s.listener.Accept()
			if s.net == "kcp" {
				if e != nil {
					log.Error("accept failed: ", e)
					continue
				}

				log.Info("kcp accept ", conn.RemoteAddr(), e)

				udpConn := conn.(*kcp.UDPSession)
				//udpConn.SetWindowSize(2048, 2048)
				udpConn.SetWindowSize(1024, 2048)
				udpConn.SetMtu(900)
				udpConn.SetNoDelay(1, 10, 2, 1)
				udpConn.SetWriteDelay(false)
				udpConn.SetACKNoDelay(true)
			} else {
				if e != nil {
					if ne, ok := e.(net.Error); ok && ne.Temporary() {
						if tempDelay == 0 {
							tempDelay = 5 * time.Millisecond
						} else {
							tempDelay *= 2
						}
						if max := 1 * time.Second; tempDelay > max {
							tempDelay = max
						}
						log.Error("http: Accept error: %v; retrying in %v", e, tempDelay)
						time.Sleep(tempDelay)
						continue
					}
					break
				}
				tcpConn := conn.(*net.TCPConn)
				tcpConn.SetNoDelay(true)
			}
			tempDelay = 0
			id := s.nextID()
			session := newSession(id, conn, s.proto)
			s.eq <- newEvent(EVENT_ACCEPT, session, nil)
		}
	}()
	return nil
}

func (s *Server) nextID() int64 {
	return <-s.idpool
}

func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
