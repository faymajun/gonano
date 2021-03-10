package anet

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	MAXN_BUFFER_SIZE = 65535
)

type UDPSession struct {
	id     int64
	conn   net.PacketConn
	proto  Protocol
	wbuf   chan Message
	events chan Event
	ctrl   chan struct{}
	net    string
}

func newUDPSession(id int64, conn net.PacketConn, proto Protocol) *UDPSession {
	sess := &UDPSession{
		id:     id,
		conn:   conn,
		proto:  proto,
		wbuf:   make(chan Message, SEND_BUFF_SIZE),
		events: nil,
		ctrl:   make(chan struct{}, 1),
		net:    "",
	}
	return sess
}

func (self *UDPSession) Start(events chan Event) {
	if events != nil {
		self.events = events
	}
	go self.reader()
	go self.writer()
}

func (self *UDPSession) ID() int64 {
	return self.id
}

func (self *UDPSession) Close() {
	self.conn.Close()
}

func (self *UDPSession) Send(api uint8, payload interface{}) {
}

func (self *UDPSession) SendPeer(peer net.Addr, api uint8, payload interface{}) {
	defer func() {
		if x := recover(); x != nil {
			log.Infof("Send Error: %s", x)
		}
	}()
	//log.Info("session: ", self.ID(), " try to send pkt: ", api)

	if len(self.wbuf) < SEND_BUFF_SIZE {
		self.wbuf <- Message{peer, api, payload}
	} else {
		log.Info("send overflow ", self.ID())
	}
}

func (self *UDPSession) reader() {
	log.Infof("session[%d] start reader...", self.id)
	defer func() {
		log.Infof("reader[%d] quit...", self.id)
		self.ctrl <- struct{}{}
		self.events <- newEvent(EVENT_DISCONNECT, self, nil)
	}()
	for {
		buf := make([]byte, MAXN_BUFFER_SIZE)
		n, addr, err := self.conn.ReadFrom(buf)
		if err != nil {
			log.Error("read packet failed: ", err)
			break
		}
		log.Info("recv packet size: ", n)
		log.Info("packet: ", buf[0:n])

		header := buf[0:4]
		size := binary.LittleEndian.Uint32(header) - 4
		if size > MAXN_PACKET_SIZE {
			log.Error("invalid package size: ", size)
			continue
		}
		api, payload, err := self.proto.Decode(buf[4:n])
		if err != nil {
			log.Error("invalid package type: ", addr.String(), ", ", err)
			self.events <- newEvent(EVENT_RECV_ERROR, self, err)
		} else {
			msg := NewMessage(addr, api, payload)
			self.events <- newEvent(EVENT_MESSAGE, self, msg)
		}
	}
}

func (self *UDPSession) writer() {
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
					if n, err := self.conn.WriteTo(raw, msg.Peer); err != nil {
						log.Error("raw send error ", err, " ", msg.Peer, " size:", len(raw))
						self.events <- newEvent(EVENT_SEND_ERROR, self, err)
					} else {
						log.Info("send success: ", msg.Peer, ", size: ", n)
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
