package group

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/faymajun/gonano/message"

	"github.com/faymajun/gonano/colorized"
	"github.com/faymajun/gonano/tags"

	"fmt"

	"github.com/faymajun/gonano/core"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

const (
	statusWorking = 0
	statusClosed  = 1
)

var logger = logrus.WithField("component", "group")

var (
	ErrCloseClosedGroup   = errors.New("close closed group")
	ErrClosedGroup        = errors.New("group closed")
	ErrSessionDuplication = errors.New("session has existed in the current group")
	ErrSessionNull        = errors.New("session is nil")
	ErrSessionNoBind      = errors.New("session is not bind")
)

// SessionFilter represents a filter which was used to filter session when Multicast,
// the session will receive the message while filter returns true.
type SessionFilter func(*core.Session) bool

// Group represents a session group which used to manage a number of
// sessions, data send to the group will send to all session in it.
type Group struct {
	sync.RWMutex
	status   int32                           // channel current status
	name     string                          // channel name
	sessions map[*core.Session]*core.Session // session id map to session instance

	muBuf  sync.Mutex    // protect buffer
	buffer *proto.Buffer // proto marshal buffer
}

// NewGroup returns a new group instance
func New(name string) *Group {
	return &Group{
		status:   statusWorking,
		name:     name,
		sessions: make(map[*core.Session]*core.Session),
		buffer:   proto.NewBuffer(nil),
	}
}

func (c *Group) marshal(v proto.Message) ([]byte, error) {
	c.muBuf.Lock()
	defer c.muBuf.Unlock()

	c.buffer.Reset()
	err := c.buffer.Marshal(v)
	if err != nil {
		return nil, err
	}
	buf := c.buffer.Bytes()
	ret := make([]byte, len(buf))
	copy(ret, buf)
	return ret, nil
}

// Push message to partial client, which filter return true
func (c *Group) Multicast(msgid message.MSGID, v proto.Message, filter SessionFilter) error {
	if c.isClosed() {
		return ErrClosedGroup
	}

	if tags.DEBUG {
		println(fmt.Sprintf(colorized.Magenta("==>>>> %s, %+v"), msgid.String(), v))
	}

	// 这里特殊处理，在主线程序列化消息，可以有效减少每个用户线程都去序列化消息
	data, err := c.marshal(v)
	if err != nil {
		logger.Error(err)
		return err
	}

	c.RLock()
	defer c.RUnlock()

	for s := range c.sessions {
		if !filter(s) {
			continue
		}
		if err := s.SendData(msgid, data); err != nil {
			//logger.Warnf("Group.Multicast: RemoteAddr=%s, Error=%s, ", s.RemoteAddr().String(), err.Error())
		}
	}

	return nil
}

// Push message to all client
func (c *Group) Broadcast(msgid message.MSGID, v proto.Message) error {
	return c.Multicast(msgid, v, func(cSession *core.Session) bool {
		return true
	})
}

func (c *Group) BroadcastData(msgid message.MSGID, data []byte) error {
	buf := make([]byte, len(data))
	copy(buf, data)

	c.RLock()
	defer c.RUnlock()

	for s := range c.sessions {
		if s != nil {
			if err := s.SendData(msgid, buf); err != nil {
				//logger.Warnf("Group.BroadcastData: RemoteAddr=%s, Error=%s, ", s.RemoteAddr().String(), err.Error())
			}
		}

	}
	return nil
}

//func (c *Group) RelateBroadcastData(msgid message.MSGID, data []byte) error {
//	buf := net.BattleUdpCodecFactory().Encode(msgid, data)
//
//	c.RLock()
//	defer c.RUnlock()
//
//	for _, v := range c.sessions {
//		if v != nil {
//			// 先序列化编码以后再发送。避免底层多次序列化编码
//			if err := v.SendDataNotEncode(msgid, buf); err != nil {
//				//logger.Warnf("Group.RelateBroadcastData: RemoteAddr=%s, Error=%s, ", s.RemoteAddr().String(), err.Error())
//			}
//		}
//	}
//
//	return nil
//}

// Add add session to group
func (c *Group) Add(session *core.Session) error {
	if session == nil {
		return ErrSessionNull
	}

	if c.isClosed() {
		return ErrClosedGroup
	}

	c.Lock()
	defer c.Unlock()

	if nil == session {
		return nil
	}

	_, ok := c.sessions[session]
	if ok {
		return ErrSessionDuplication
	}

	c.sessions[session] = nil
	return nil
}

func (c *Group) CheckSess(session *core.Session, val *core.Session) error {
	if session == nil || val == nil {
		return ErrSessionNull
	}

	if c.isClosed() {
		return ErrClosedGroup
	}

	if _, ok := c.sessions[session]; !ok {
		return ErrSessionNoBind
	}

	if c.sessions[session] == val {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	if c.sessions[session] == val {
		return nil
	}

	// 关闭旧连接
	if c.sessions[session] != nil {
		c.sessions[session].Close()
	}

	// 关联新连接
	c.sessions[session] = val
	return nil
}

// Leave remove specified UID related session from group
func (c *Group) Leave(session *core.Session) error {
	if session == nil {
		return ErrSessionNull
	}

	if c.isClosed() {
		return ErrClosedGroup
	}

	session.Close()
	if v, ok := c.sessions[session]; ok {
		if v != nil {
			v.Close()
		}
	}

	c.Lock()
	defer c.Unlock()

	delete(c.sessions, session)
	return nil
}

// LeaveAll clear all sessions in the group
func (c *Group) LeaveAll() error {
	if c.isClosed() {
		return ErrClosedGroup
	}

	for k, v := range c.sessions {
		if k != nil {
			k.Close()
		}
		if v != nil {
			v.Close()
		}
	}

	c.sessions = make(map[*core.Session]*core.Session)
	return nil
}

// Count get current member amount in the group
func (c *Group) Count() int {
	c.RLock()
	defer c.RUnlock()

	return len(c.sessions)
}

func (c *Group) isClosed() bool {
	if atomic.LoadInt32(&c.status) == statusClosed {
		return true
	}
	return false
}

// Close destroy group, which will release all resource in the group
func (c *Group) Close() error {
	if c.isClosed() {
		return ErrCloseClosedGroup
	}

	c.LeaveAll()
	atomic.StoreInt32(&c.status, statusClosed)
	return nil
}
