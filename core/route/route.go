package route

import (
	"reflect"

	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// 消息ID对应的回调处理函数
var handlers = [message.MSGID_MAX_COUNT]LogicHandler{}

// 注册消息回调函数
func register(msgid message.MSGID, handler HandlerFunc, payload proto.Message, isRaw ...bool) {
	if int(msgid) >= len(handlers) {
		log.Errorf("消息ID超出范围: 最大值=%d, 当前值=%d, ID=%s", len(handlers), int(msgid), msgid.String())
		return
	}

	if handler == nil {
		log.Error("消息处理函数不可为空")
		return
	}

	// 检查是都已经注册
	if h := handlers[msgid]; h.fn != nil {
		log.Errorf("消息ID已经注册: %s", msgid.String())
		return
	}

	keepRaw := len(isRaw) > 0 && isRaw[0]
	payloadType := reflect.TypeOf(payload)
	if payloadType.Kind() != reflect.Ptr {
		log.Errorf("消息类型必须为一个指针: %s", msgid.String())
		return
	}
	handlers[msgid] = LogicHandler{msgid: msgid, fn: handler, typ: payloadType.Elem(), raw: keepRaw}
	log.Infof("注册消息: MsgID=%s", msgid)
}

func Register(msgid message.MSGID, handler HandlerFunc, payload proto.Message) {
	register(msgid, handler, payload)
}

// 注册消息处理函数
func RegisterTable(table []Stub) {
	for _, item := range table {
		register(item.MsgID, item.Handler, item.Payload)
	}
}

// 注册转发消息和透传函数
func RegisterForward(table []Stub, handlerFunc HandlerFunc) {
	for _, item := range table {
		register(item.MsgID, handlerFunc, item.Payload, true)
	}
}

func UnRegister(msgid message.MSGID) {
	if int(msgid) >= len(handlers) {
		log.Errorf("消息ID大于65535 %s", msgid.String())
		return
	}

	handlers[msgid] = LogicHandler{}
}

// 获取消息ID对应的处理器函数
// msgid:   消息ID
//
// 返回值
// receive: 处理器
// error:   错误, 如果handler没有找到
func FindHandler(msgid message.MSGID) (*LogicHandler, error) {
	if int(msgid) >= len(handlers) {
		return nil, errors.Errorf("message id over range: %s", msgid.String())
	}

	if h := handlers[msgid]; h.fn == nil {
		return nil, errors.Errorf("receive not found: %s", msgid.String())
	}

	return &handlers[msgid], nil
}

func NewTestHandler(msgid message.MSGID) (*LogicHandler, error) {
	return &LogicHandler{msgid: msgid}, nil
}
