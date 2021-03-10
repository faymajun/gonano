package anet

import (
	"net"
)

const (
	EVENT_ACCEPT          = 0x01
	EVENT_CONNECT_SUCCESS = 0x02
	EVENT_CONNECT_FAILED  = 0x04
	EVENT_DISCONNECT      = 0x08
	EVENT_MESSAGE         = 0x10
	EVENT_RECV_ERROR      = 0x20
	EVENT_SEND_ERROR      = 0x40
)

type BaseSession interface {
	Start(events chan Event)
	ID() int64
	Close()
	Send(api uint8, payload interface{})
	SendPeer(peer net.Addr, api uint8, payload interface{})
}

type Event struct {
	Type    int8
	Session BaseSession
	Data    interface{}
}

func newEvent(typ int8, session BaseSession, data interface{}) Event {
	return Event{
		Type:    typ,
		Session: session,
		Data:    data,
	}
}
