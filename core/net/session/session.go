package session

import (
	"github.com/faymajun/gonano/message"
	"net"

	"github.com/golang/protobuf/proto"
)

type (
	sender interface {
		RemoteAddr() net.Addr
		Send(msgid message.MSGID, pbmsg proto.Message) error
		SendData(msgid message.MSGID, data []byte) error
		SendDataNotEncode(msgid message.MSGID, data []byte) error
		Close() error
		SessionID() uint32
	}

	Entity interface {
		ID() int64
	}
)

type Session struct {
	sender        // 底层网络抽象
	entity Entity // 不同服务器, 绑定不同的网络实体
}

func NewSession(sender sender) *Session {
	s := &Session{
		sender: sender,
	}
	return s
}

func (s *Session) SetEntity(entity Entity) { s.entity = entity }
func (s *Session) HasEntity() bool         { return s.entity != nil }
func (s *Session) Entity() Entity          { return s.entity }
func (s *Session) Sender() sender          { return s.sender }
func (s *Session) Close() error {
	s.entity = nil
	return s.sender.Close()
}
