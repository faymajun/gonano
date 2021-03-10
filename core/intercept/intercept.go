package intercept

import (
	"github.com/faymajun/gonano/core/net/session"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type (
	Stub struct {
		MsgID       message.MSGID
		Interceptor Func
	}

	Func func(session *session.Session, data proto.Message) (receiveId string, err error)
)

var intercepts = [message.MSGID_MAX_COUNT]Func{}

func Register(msgid message.MSGID, interceptor Func) {
	if int(msgid) >= len(intercepts) {
		panic("message id exceed")
		return
	}

	if interceptor == nil {
		panic("register nil interceptor")
		return
	}

	// 检查是都已经注册
	if h := intercepts[msgid]; h != nil {
		panic("message id has been registered")
		return
	}

	intercepts[msgid] = interceptor
}

// 注册消息处理函数
func RegisterTable(table []Stub) {
	for _, item := range table {
		Register(item.MsgID, item.Interceptor)
	}
}

func Find(msgid message.MSGID) (Func, error) {
	if int(msgid) >= len(intercepts) {
		return nil, errors.Errorf("intercepts message id over range: %s", msgid.String())
	}

	if h := intercepts[msgid]; h == nil {
		return nil, errors.Errorf("intercepts receive not found: %s", msgid.String())
	}

	return intercepts[msgid], nil
}
