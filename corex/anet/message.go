package anet

import (
	"net"
)

type Message struct {
	Peer    net.Addr
	Api     uint8
	Payload interface{}
}

func NewMessage(peer net.Addr, api uint8, payload interface{}) *Message {
	return &Message{
		Peer:    peer,
		Api:     api,
		Payload: payload,
	}
}
