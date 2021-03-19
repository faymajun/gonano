package net

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/faymajun/gonano/core/coroutine"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/colorized"
	"github.com/faymajun/gonano/core/net/session"
	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/core/scheduler"
	"github.com/faymajun/gonano/message"
	"github.com/faymajun/gonano/tags"

	"github.com/golang/protobuf/proto"
)

var sessionId uint32
var sessions sync.Map

func StopTcpSession() {
	logger.Infof("<<<sessions is stop start>>>")
	sessions.Range(func(k, v interface{}) bool {
		if s, ok := v.(*TCPSession); ok {
			s.Close()
		}
		return true
	})
	logger.Infof("<<<sessions is stop over>>>")
}

func SessionsFuncRange(fn func(k, v interface{}) bool) {
	sessions.Range(fn)
}

func GetSession(sid uint32) *TCPSession {
	v, ok := sessions.Load(sid)
	if ok {
		return v.(*TCPSession)
	} else {
		return nil
	}
}

const (
	WriteQueLen = 64 //写队列长度
	ReadQueLen  = 64 //收队列长度

	bufferSize = 512 // TCP接收缓冲区大小
)

const (
	TCP_INIT = iota //待初始化
	TCP_AVAI        //可用
	TCP_STOP        //停止
)

const (
	disableResDebugDelay = true
)

type TCPSession struct {
	conn    net.Conn
	chWrite chan *packet.SendMessage // 写通道
	breader *bufio.Reader            // 读缓冲
	state   int32                    // 状态
	sid     uint32                   // 唯一标识
	coder   Codec                    // 消息编解码
	client  *TCPClient

	userSession *session.Session
	handler     SessionHandler //接受到一个消息包给上层处理
	maxRecvSize uint32         //单个包最大收包大小(<=0无限制)
	maxRecvNum  uint16         //每秒最大收包数量(<=0无限制)
	curRecvTime int64          //收包时间
	curRecvNum  uint16         //当前收包数量
	isProcess   bool
	routine     *coroutine.Coroutine //process 协程
}

func newTcpSession(conn net.Conn, coder Codec, handler SessionHandler, maxRecvSize uint32, maxRecvNum uint16, isProcess bool, wQueLen, rQueLen int) *TCPSession {
	tcpSession := &TCPSession{
		conn:    conn,
		chWrite: make(chan *packet.SendMessage, wQueLen),
		breader: bufio.NewReader(conn),
		state:   TCP_AVAI,
		sid:     atomic.AddUint32(&sessionId, 1),
		coder:   coder,
		handler: handler,

		maxRecvSize: maxRecvSize,
		maxRecvNum:  maxRecvNum,
		routine:     nil,
	}
	tcpSession.userSession = session.NewSession(tcpSession)
	sessions.Store(tcpSession.sid, tcpSession)
	tcpSession.isProcess = isProcess

	if tcpSession.isProcess {
		routine := coroutine.NewCoroutine(rQueLen, int64(tcpSession.sid))
		tcpSession.SetRoutine(routine)
	}

	return tcpSession
}

func (s *TCPSession) SessionID() uint32 {
	return s.sid
}

func (s *TCPSession) TCPClient() *TCPClient {
	return s.client
}

func (s *TCPSession) SetRoutine(routine *coroutine.Coroutine) {
	s.routine = routine
}

func (s *TCPSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }

func (s *TCPSession) Close() error {
	if !atomic.CompareAndSwapInt32(&s.state, TCP_AVAI, TCP_STOP) {
		return ErrDupClosed
	}

	sessions.Delete(s.sid)
	close(s.chWrite)

	if s.routine != nil {
		s.routine.PushTask(func() {
			s.handler.OnSessionClose(s.userSession)
		}, false)

		if s.isProcess {
			s.routine.Close()
		}
		s.routine = nil
	} else {
		scheduler.PushTask(func() { s.handler.OnSessionClose(s.userSession) })
	}

	return s.conn.Close()
}

func (s *TCPSession) isConnected() bool {
	return atomic.LoadInt32(&s.state) == TCP_AVAI
}

func (s *TCPSession) write() {
loop:
	for {
		select {
		case msg, ok := <-s.chWrite:
			if !ok {
				logger.Debugf("session:%d write is closed", s.sid)
				break loop
			}
			payload, err := msg.Serialize()
			if err != nil {
				logger.Errorf("session:%d serialize msg:%d failed:%s", s.sid, msg.MsgID, err)
				break loop
			}
			data := s.coder.Encode(msg.MsgID, payload)

			_, errWrite := s.conn.Write(data)
			if errWrite != nil {
				logger.Errorf("session:%d  write err:%v", s.sid, errWrite)
				break loop
			}
		}
	}
	s.Close()
}

func (s *TCPSession) read(heartbeat time.Duration) {
	var head [MsgHeadSize]byte
	content := make([]byte, bufferSize)

	for {
		s.conn.SetReadDeadline(time.Now().Add(time.Second * heartbeat))
		if _, err := io.ReadFull(s.breader, head[:]); err != nil {
			if err != io.EOF {
				logger.Warnf("session:%d read head is not io.EOF err:%s", s.sid, err)
			}

			logger.Debugf("session:%d read is closed. err=%v", s.sid, err)
			break
		}

		size := binary.LittleEndian.Uint32(head[:]) - 4
		if size > s.maxRecvSize && s.maxRecvSize > 0 {
			logger.Warnf("session:%d size too big:%d, addr=%s", s.sid, size, s.RemoteAddr())
			break
		}
		if s.curRecvNum > s.maxRecvNum && s.maxRecvNum > 0 {
			logger.Warnf("session:%d recv too many msg:%d, addr=%s", s.sid, s.curRecvNum, s.RemoteAddr())
			break
		}

		if size < MsgIdSize {
			logger.Errorf("session:%d size too small:%d, addr=%s", s.sid, size, s.RemoteAddr())
			break
		}

		if size > uint32(len(content)) {
			content = make([]byte, size)
		}

		if _, err := io.ReadFull(s.breader, content[:size]); err != nil {
			if err != io.EOF {
				logger.Warnf("session:%d read data err:%s, addr=%s", s.sid, err, s.RemoteAddr())
			}
			break
		}

		pack, err := s.coder.Decode(content[:size])
		if err != nil {
			logger.Warnf("session:%d decode data err:%s, addr=%s", s.sid, err, s.RemoteAddr())
			break
		}

		if pack == packet.EmptyRecvPack {
			continue
		}

		if s.curRecvTime != time.Now().Unix() {
			s.curRecvTime = time.Now().Unix()
			s.curRecvNum = 0
		}
		s.curRecvNum++
		pack.Session = s.userSession

		// 心跳直接回复
		if msg, ok := pack.Payload.(*message.ReqHeartbeat); ok {
			response := &message.ResHeartbeat{Uid: msg.Uid, ServerUnixTime: time.Now().Unix() * 1000}
			s.Send(message.MSGID_ResHeartbeatE, response)
			continue
		}

		//// 登录和重连前面登录验证 到主线程
		//if _, ok := pack.Payload.(*message.Growth_ReqRoleLogin); ok {
		//	scheduler.PushPacket(&pack)
		//	continue
		//}
		//
		//if _, ok := pack.Payload.(*message.Growth_ReqRoleRecon); ok {
		//	scheduler.PushPacket(&pack)
		//	continue
		//}

		if s.routine != nil {
			// 自身逻辑线程处理
			s.routine.PushPacket(&pack)
		} else {
			// 主线程处理
			scheduler.PushPacket(&pack)
		}
	}
	s.Close()
}

func (s *TCPSession) send(msg *packet.SendMessage) error {
	if !s.isConnected() {
		return fmt.Errorf("session is closed: %s", s.RemoteAddr().String())
	}

	if len(s.chWrite) >= cap(s.chWrite) {
		logger.Warnf("send buffer excced, session close", s.RemoteAddr())

		// 立即断开，走重连的流程
		s.Close()
		return fmt.Errorf("send buffer excced: %s", s.RemoteAddr())
	}

	if tags.DEBUG {
		if msg.MsgID != message.MSGID_ResHeartbeatE &&
			msg.MsgID != message.MSGID_ReqHeartbeatE {
			println(fmt.Sprintf(colorized.Magenta("==> %s, %+v, %s"), msg.MsgID.String(), msg.Payload, s.RemoteAddr()))
		}
	}

	routine.Try(func() { s.chWrite <- msg }, nil)
	return nil
}

func (s *TCPSession) Send(msgid message.MSGID, pbmsg proto.Message) error {
	return s.send(&packet.SendMessage{MsgID: msgid, Payload: pbmsg})
}

func (s *TCPSession) SendData(msgid message.MSGID, data []byte) error {
	return s.send(&packet.SendMessage{MsgID: msgid, Payload: data})
}

func (s *TCPSession) SendDataNotEncode(msgid message.MSGID, data []byte) error {
	// 暂时未实现
	return nil
}

func (s *TCPSession) UserSession() *session.Session {
	return s.userSession
}
