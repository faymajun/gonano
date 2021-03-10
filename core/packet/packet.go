package packet

import (
	"errors"

	"github.com/faymajun/gonano/core/net/session"
	"github.com/faymajun/gonano/core/route"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
)

var (
	ErrUnsupportedPayload = errors.New("发送数据包不支持的Payload类型")
)

type (
	SendMessage struct {
		MsgID message.MSGID
		// 消息可以是proto.Message或[]byte, 如
		// 果是[]byte表示广播消息, 已经提前序列化好
		Payload    interface{}
		DontEncode bool //是否需要encode, 默认为false,需要encode
	}

	RecvMessage struct {
		Session *session.Session
		Handler *route.LogicHandler
		Payload proto.Message
	}
)

var EmptyRecvPack = RecvMessage{}

// 返回序列化后的数据
func (p *SendMessage) Serialize() ([]byte, error) {
	switch v := p.Payload.(type) {
	case []byte:
		return v, nil
	case proto.Message:
		return proto.Marshal(v)
	default:
		return nil, ErrUnsupportedPayload
	}
}
