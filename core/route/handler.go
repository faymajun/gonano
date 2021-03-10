/*
	日期: 2017-07-06
	作者: Long Heng
	功能: 消息处理函数包装类, 类中包含回调函数和方法签名的第二个类型
*/
package route

import (
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/faymajun/gonano/core/net/session"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type (
	Stub struct {
		MsgID   message.MSGID
		Handler HandlerFunc
		Payload proto.Message
	}

	// 消息处理函数
	HandlerFunc func(*session.Session, proto.Message) error

	// 消息处理器
	LogicHandler struct {
		msgid message.MSGID // 消息ID
		fn    HandlerFunc   // 回调函数
		typ   reflect.Type  // 消息类型
		raw   bool
	}
)

// 使用反序列化好的消息和玩家对应的session调用消息处理函数
func (h *LogicHandler) Handle(s *session.Session, payload proto.Message) error {
	// 防止逻辑处理过程中panic导致逻辑线程crash
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Handler panic: MsgID=%s, Error=%v", h.msgid, err)
			fmt.Fprintln(logrus.StandardLogger().Out, string(debug.Stack()))
		}
	}()

	// logic processor
	return h.fn(s, payload)
}

// 生成一个新的消息实例
func (h *LogicHandler) Instance() proto.Message {
	return reflect.New(h.typ).Interface().(proto.Message)
}

func (h *LogicHandler) MsgID() message.MSGID {
	return h.msgid
}

func (h *LogicHandler) IsRaw() bool {
	return h.raw
}
