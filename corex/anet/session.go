package anet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Session struct {
	id            int64
	conn          net.Conn
	proto         Protocol
	wbuf          chan Message
	events        chan Event
	ctrl          chan struct{}
	net           string
	raddr         *net.TCPAddr
	autoReconnect bool
	reconnect     chan bool
}

const (
	SEND_BUFF_SIZE   = 65535
	CONNECT_INTERVAL = 1000             // reconnect interval
	READ_TIMEOUT     = 600              // 5 min
	WRITE_TIMEOUT    = 300              // 5 min
	MAXN_PACKET_SIZE = 10 * 1024 * 1024 // 5MB
)

func newSession(id int64, conn net.Conn, proto Protocol) *Session {
	sess := &Session{
		id:            id,
		conn:          conn,
		proto:         proto,
		wbuf:          make(chan Message, SEND_BUFF_SIZE),
		events:        nil,
		ctrl:          make(chan struct{}, 1),
		net:           "",
		raddr:         nil,
		autoReconnect: false,
		reconnect:     nil,
	}
	return sess
}

func ConnectTo(network string, addr string, proto Protocol, events chan Event, autoReconnect bool) *Session {
	session := newSession(0, nil, proto)
	session.connect(network, addr, events, autoReconnect)
	return session
}

func (self *Session) Start(events chan Event) {
	if events != nil {
		self.events = events
	}
	go self.reader()
	go self.writer()
}

func (self *Session) ID() int64 {
	return self.id
}

func (self *Session) Close() {
	if self.autoReconnect {
		self.reconnect <- false
	}
	self.conn.Close()
}

func (self *Session) Send(api uint8, payload interface{}) {
	defer func() {
		if x := recover(); x != nil {
			log.Infof("Send Error: %s", x)
		}
	}()
	//log.Info("session: ", self.ID(), " try to send pkt: ", api)

	if len(self.wbuf) < SEND_BUFF_SIZE {
		self.wbuf <- Message{nil, api, payload}
	} else {
		log.Info("send overflow ", self.ID())
	}
}

func (self *Session) reader() {
	log.Infof("session[%d] start reader...", self.id)
	defer func() {
		log.Infof("reader[%d] quit...", self.id)
		self.ctrl <- struct{}{}
		if self.autoReconnect {
			self.reconnect <- true
		} else {
			self.events <- newEvent(EVENT_DISCONNECT, self, nil)
		}
	}()
	header := make([]byte, 4)
	for {
		//self.conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT * time.Second))
		if _, err := io.ReadFull(self.conn, header); err != nil {
			log.Error("readfull failed: ", err)
			break
		}

		log.Info("header: ", header)

		size := binary.LittleEndian.Uint32(header) - 4
		if size > MAXN_PACKET_SIZE {
			log.Error("invalid package size: ", size, " ", self.conn.RemoteAddr().String())
			break
		}
		data := make([]byte, size)
		//self.conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT * time.Second))
		if _, err := io.ReadFull(self.conn, data); err != nil {
			log.Infof("io.ReadFull() error: %s", err)
			self.events <- newEvent(EVENT_RECV_ERROR, self, err)
			break
		}

		api, payload, err := self.proto.Decode(data)
		if err != nil {
			log.Error("invalid package type: ", self.conn.RemoteAddr().String(), ", ", err)
			self.events <- newEvent(EVENT_RECV_ERROR, self, err)
			break
		} else {
			msg := NewMessage(self.conn.RemoteAddr(), api, payload)
			self.events <- newEvent(EVENT_MESSAGE, self, msg)
		}
	}
}

func encode(proto Protocol, msg Message) ([]byte, error) {
	data, err := proto.Encode(msg.Api, msg.Payload)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(data)+4)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	//log.Infof("send data size: %d, type=%d", len(data), msg.Api)
	//log.Info("playload: ", buf.Bytes())
	return buf.Bytes(), nil
}

func rawSend(w *bufio.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return err
	}
	return w.Flush()
}

func (self *Session) writer() {
	log.Infof("session[%d] start writer...", self.id)
	defer func() {
		log.Infof("writer[%d] quit ...", self.id)
		close(self.wbuf)
		self.conn.Close()
		self.wbuf = nil

		///
		if x := recover(); x != nil {
			log.Infof("Send Error: %s", x)
		}
	}()

	w := bufio.NewWriter(self.conn)
	for {
		select {
		case msg, ok := <-self.wbuf:
			if ok {
				if raw, err := encode(self.proto, msg); err != nil {
					self.events <- newEvent(EVENT_SEND_ERROR, self, err)
					log.Error("encode msg error ", msg, " ", err, " ", self.id)
					return
				} else {
					self.conn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT * time.Second))
					if err := rawSend(w, raw); err != nil {
						self.events <- newEvent(EVENT_SEND_ERROR, self, err)
						log.Error("raw send error ", err, " ", self.id, " size:", len(raw))
						return
					}
					self.conn.SetWriteDeadline(time.Time{})
				}
			} else {
				log.Error("get ev from wbuf failed ", self.id)
				return
			}
		case <-self.ctrl:
			return
		}
	}
}

func (self *Session) supervisor() {
	defer func() {
		log.Info("supervisor quit...")
	}()
	for {
		select {
		case flag, ok := <-self.reconnect:
			if ok {
				if flag {
					log.Infof("reconnect to %s", self.raddr)
					go self.connector()
				} else {
					return
				}
			}
		}
	}
}

func (self *Session) connect(network string, addr string, events chan Event, autoReconnect bool) error {
	log.Infof("try to connect to %s %s", network, addr)
	raddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		log.Infof("net.ResolveTCPAddr: error: %s", err)
		return err
	}
	self.events = events
	self.net = network
	self.raddr = raddr
	if autoReconnect {
		self.autoReconnect = autoReconnect
		self.reconnect = make(chan bool, 1)
		go self.supervisor()
	}
	go self.connector()
	return nil
}

func (self *Session) connector() {
	conn, err := net.DialTCP(self.net, nil, self.raddr)
	if err != nil {
		log.Infof("connect to %s falied: %s, id=%d", self.raddr, err, self.id)
		if self.autoReconnect {
			time.Sleep(CONNECT_INTERVAL * time.Millisecond)
			self.reconnect <- true
		} else {
			self.events <- newEvent(EVENT_CONNECT_SUCCESS, self, err)
		}
	} else {
		log.Infof("connect to %s ok...id=%d", self.raddr, self.id)
		self.conn = conn
		if !self.autoReconnect {
			self.events <- newEvent(EVENT_CONNECT_SUCCESS, self, nil)
		} else {
			self.wbuf = make(chan Message, SEND_BUFF_SIZE)
			self.Start(self.events)
		}
	}
}

func (self *Session) RemoteAddr() string {
	if self.raddr == nil {
		return ""
	}
	return self.raddr.String()
}

func (self *Session) SendPeer(peer net.Addr, api uint8, payload interface{}) {
	log.Info("not implement...")
}
