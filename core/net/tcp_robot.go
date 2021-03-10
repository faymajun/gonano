package net

//
//import (
//	"fmt"
//	"net"
//	"reflect"
//	"github.com/faymajun/gonano/config"
//	"github.com/faymajun/gonano/core/coroutine"
//	"sync"
//	"sync/atomic"
//	"time"
//
//	"github.com/faymajun/gonano/core"
//	"github.com/faymajun/gonano/core/packet"
//	"github.com/faymajun/gonano/core/routine"
//	"github.com/faymajun/gonano/core/scheduler"
//	"github.com/faymajun/gonano/message"
//
//	"github.com/golang/protobuf/proto"
//	log "github.com/sirupsen/logrus"
//)
//
//type TCPRobot struct {
//	session      *TCPSession    // Session
//	state        int32          // 状态
//	addr         string         // 对端地址
//	receive      SessionHandler // 消息处理器
//	cid          uint32         // 唯一标识
//	codecFactory func() Codec   // 消息编解码
//	heartClose   chan struct{}  // 心跳退出信号
//	msgs         map[message.MSGID][]*packet.RecvMessage
//	mutex        sync.Mutex
//	chRead       chan *packet.RecvMessage // 读通道
//	chFunc       chan func()              // 处理函数channel
//	rid          int64                    //
//	teamId       string
//}
//
//func (client *TCPRobot) Close() error {
//	if !atomic.CompareAndSwapInt32(&client.state, TCP_AVAI, TCP_STOP) {
//		return ErrDupClosed
//	}
//
//	clients.Delete(client.cid)
//	close(client.heartClose)
//
//	if client.session != nil {
//		client.session.Close()
//	}
//
//	close(client.chFunc)
//	close(client.chRead)
//	return nil
//}
//
//func (client *TCPRobot) SetID(Id int64) {
//	client.rid = Id
//}
//
//func (client *TCPRobot) ID() int64 {
//	return client.rid
//}
//
//func (client *TCPRobot) SetTeamID(Id string) {
//	client.teamId = Id
//}
//
//func (client *TCPRobot) TeamID() string {
//	return client.teamId
//}
//
//func (client *TCPRobot) isRunning() bool {
//	return atomic.LoadInt32(&client.state) == TCP_AVAI
//}
//
//func StartTCPRobot(addr string, receive SessionHandler, codecFactory func() Codec) (*TCPRobot, error) {
//	if receive == nil {
//		return nil, fmt.Errorf("session receive can not be nil")
//	}
//
//	client := &TCPRobot{
//		state:        TCP_AVAI,
//		addr:         addr,
//		receive:      receive,
//		cid:          atomic.AddUint32(&clientId, 1),
//		codecFactory: codecFactory,
//		heartClose:   make(chan struct{}),
//		msgs:         make(map[message.MSGID][]*packet.RecvMessage),
//		chRead:       make(chan *packet.RecvMessage, 204800),
//		chFunc:       make(chan func(), 204800),
//	}
//
//	clients.Store(client.cid, client)
//	client.connect()
//	return client, nil
//}
//
//func (client *TCPRobot) connect() {
//	_, err := client.dial()
//	for err != nil {
//		fmt.Println("connect err ", err)
//		time.Sleep(time.Second * ReconnectInterval)
//		if !client.isRunning() {
//			break
//		}
//		_, err = client.dial()
//	}
//}
//
//func (client *TCPRobot) dial() (conn net.Conn, err error) {
//	conn, err = net.DialTimeout("tcp4", client.addr, time.Second*DialTimeout)
//	if err != nil {
//		return
//	}
//
//	tcpSession := newTcpSession(conn, client.codecFactory(), client.receive, 0, 0, false, 204800, 204800)
//	client.session = tcpSession
//	scheduler.PushTask(func() { client.receive.OnSessionCreate(tcpSession.userSession) })
//	routine.Go(func() { client.session.read(ReadTimeOut) })
//	routine.Go(func() { client.session.write() })
//	routine.Go(func() { client.heartbeat() })
//	rt := coroutine.TestCoroutine(client.chRead, client.chFunc)
//	tcpSession.SetRoutine(rt)
//	routine.Go(func() {
//		client.process()
//	})
//	return
//}
//
//func (c *TCPRobot) Func(f func()) {
//	c.chFunc <- f
//}
//
//func (c *TCPRobot) process() {
//loop:
//	for {
//		select {
//		case fn, ok := <-c.chFunc:
//			if !ok {
//				log.Warnf("TCPSession function queue was closed")
//				c.chFunc = nil
//				if c.chRead == nil {
//					break loop
//				}
//				continue
//			}
//			if fn == nil {
//				log.Errorf("TCPSession receive a nil function")
//				continue
//			}
//			routine.Try(fn, nil)
//
//		case pack, ok := <-c.chRead:
//			//log.Println("pack ",pack.Payload.String())
//			if !ok {
//				log.Warnf("TCPSession packets queue was closed")
//				c.chRead = nil
//				if c.chFunc == nil {
//					break loop
//				}
//				continue
//			}
//			c.mutex.Lock()
//			msgs, ok := c.msgs[pack.Handler.MsgID()]
//			if !ok {
//				msgs = make([]*packet.RecvMessage, 0)
//			}
//			msgs = append(msgs, pack)
//			c.msgs[pack.Handler.MsgID()] = msgs
//			c.mutex.Unlock()
//		}
//	}
//	c.session.Close()
//}
//
//func (client *TCPRobot) heartbeat() {
//	t := time.NewTicker(1 * time.Second)
//	p := &message.ReqHeartbeat{}
//	defer t.Stop()
//loop:
//	for {
//		select {
//		case <-client.heartClose:
//			break loop
//
//		case <-t.C:
//			if !client.isRunning() {
//				break loop
//			}
//			client.msgs = make(map[message.MSGID][]*packet.RecvMessage)
//			client.Send(message.MSGID(message.MSGID_ReqHeartbeatE), p)
//		}
//	}
//	logger.Infof("远程服务器心跳线程退出, Addr=%s", client.addr)
//}
//
//func (client *TCPRobot) Reconnect() {
//	if !client.isRunning() {
//		return
//	}
//
//	client.connect()
//}
//
//func (client *TCPRobot) Send(msgid message.MSGID, pbmsg proto.Message) error {
//	if !client.isRunning() {
//		return fmt.Errorf("client:%d is not running", client.cid)
//	}
//
//	sendTimeD := config.Content.Int("send_time_duration")
//	if sendTimeD != 0 {
//		time.Sleep(time.Duration(sendTimeD) * time.Millisecond) //防止一秒内传20个以上的包
//	}
//	sess := client.session
//	if sess != nil {
//		return sess.Send(msgid, pbmsg)
//	}
//	return fmt.Errorf("client:%d session is nil", client.cid)
//}
//
//func (client *TCPRobot) FindAll(msgid message.MSGID, payload proto.Message, isMust bool) []proto.Message {
//	client.mutex.Lock()
//	list := make([]proto.Message, 0)
//	msgs, ok := client.msgs[msgid]
//	if ok {
//		for _, pack := range msgs {
//			payloadType := reflect.TypeOf(payload)
//			pbMsg := reflect.New(payloadType.Elem()).Interface().(proto.Message)
//			pbuf := pack.Payload.(*message.BytesBuffer)
//			pbbuf := proto.NewBuffer(pbuf.Buffer)
//			if err := pbbuf.Unmarshal(pbMsg); err == nil {
//				list = append(list, pbMsg)
//			}
//		}
//	}
//	client.msgs[msgid] = nil
//	client.mutex.Unlock()
//	if len(list) == 0 {
//		if isMust {
//			time.Sleep(10 * time.Millisecond)
//			return client.FindAll(msgid, payload, isMust)
//		}
//	}
//	return list
//}
//
//func (client *TCPRobot) FindPack(msgid message.MSGID, payload proto.Message, isMust bool) proto.Message {
//	client.mutex.Lock()
//	var v proto.Message
//	msgs, ok := client.msgs[msgid]
//	if ok {
//		if len(msgs) != 0 {
//			pack := msgs[(len(msgs) - 1)]
//			pbuf := pack.Payload.(*message.BytesBuffer)
//			pbbuf := proto.NewBuffer(pbuf.Buffer)
//			if err := pbbuf.Unmarshal(payload); err == nil {
//				v = payload
//			}
//		}
//	}
//	client.msgs[msgid] = nil
//	client.mutex.Unlock()
//	if v == nil {
//		if isMust {
//			time.Sleep(10 * time.Millisecond)
//			return client.FindPack(msgid, payload, isMust)
//		}
//	}
//	return payload
//}
//
//// func (client *TCPRobot) FindPack(msgid message.MSGID, pbMsg proto.Message) proto.Message {
//// 	if !client.isRunning() {
//// 		return nil
//// 	}
//// 	for {
//// 		select {
//// 		case pack, ok := <-client.session.chRead:
//// 			if !ok {
//// 				log.Warnf("TCPSession packets queue was closed")
//// 				client.session.chRead = nil
//// 				if client.session.chFunc == nil {
//// 					return nil
//// 				}
//// 				continue
//// 			}
//// 			if pack.Handler.MsgID() == msgid {
//// 				pbuf := pack.Payload.(*message.BytesBuffer)
//// 				pbbuf := proto.NewBuffer(pbuf.Buffer)
//// 				if err := pbbuf.Unmarshal(pbMsg); err != nil {
//// 					return nil
//// 				}
//// 				return pbMsg
//// 			}
//// 		}
//// 	}
//// }
//
//func (client *TCPRobot) RemoteAddr() net.Addr {
//	return client.session.RemoteAddr()
//}
//
//func (client *TCPRobot) Addr() string {
//	return client.addr
//}
//
//func (client *TCPRobot) Session() *core.Session {
//	return client.session.userSession
//}
//
//type RoBot struct {
//	Id int64
//}
//
//func (role *RoBot) ID() int64 {
//	return role.Id
//}
//func (client *TCPRobot) SetEntity(roleId int64) {
//	role := &RoBot{Id: roleId}
//	if client.session != nil {
//		client.session.userSession.SetEntity(role)
//	}
//
//}
//func (client *TCPRobot) GetEntity() int64 {
//	if client.session == nil || client.session.userSession == nil || client.session.userSession.Entity() == nil {
//		return 0
//	}
//	return client.session.userSession.Entity().ID()
//}
