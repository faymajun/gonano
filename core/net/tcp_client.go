package net

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/core"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/core/scheduler"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
)

const (
	DialTimeout       = 5  // 连接超时
	ReconnectInterval = 10 // 重连间隔
	ReadTimeOut       = 16 // 读超时
)

var clientId uint32
var clients sync.Map
var ErrDupClosed = errors.New("duplication close")

type TCPClient struct {
	session      *TCPSession    // Session
	state        int32          // 状态
	addr         string         // 对端地址
	handler      SessionHandler // 消息处理器
	cid          uint32         // 唯一标识
	codecFactory func() Codec   // 消息编解码
	heartClose   chan struct{}  // 心跳退出信号
}

func (client *TCPClient) Close() error {
	if !atomic.CompareAndSwapInt32(&client.state, TCP_AVAI, TCP_STOP) {
		return ErrDupClosed
	}

	clients.Delete(client.cid)
	close(client.heartClose)
	if client.session == nil {
		return nil
	}
	return client.session.Close()
}

func (client *TCPClient) isRunning() bool {
	return atomic.LoadInt32(&client.state) == TCP_AVAI
}

func StartTcpClient(addr string, handler SessionHandler, codecFactory func() Codec) (*TCPClient, error) {
	if handler == nil {
		return nil, fmt.Errorf("session receive can not be nil")
	}

	client := &TCPClient{
		state:        TCP_AVAI,
		addr:         addr,
		handler:      handler,
		cid:          atomic.AddUint32(&clientId, 1),
		codecFactory: codecFactory,
		heartClose:   make(chan struct{}),
	}

	clients.Store(client.cid, client)
	client.connect()
	routine.Go(func() { client.heartbeat() })
	return client, nil
}

func (client *TCPClient) connect() {
	go func() {
		_, err := client.dial()
		for err != nil {
			time.Sleep(time.Second * ReconnectInterval)
			if !client.isRunning() {
				break
			}
			_, err = client.dial()
		}
	}()
}

const (
	WriteCap = 204800 // 20万
	ReadCap  = 204800 // 20万
)

func (client *TCPClient) dial() (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp4", client.addr, time.Second*DialTimeout)
	if err != nil {
		return
	}

	tcpSession := newTcpSession(conn, client.codecFactory(), client.handler, 0, 0, false, WriteCap, ReadCap)
	client.session = tcpSession
	tcpSession.client = client
	scheduler.PushTask(func() { client.handler.OnSessionCreate(tcpSession.userSession) })
	routine.Go(func() { client.session.read(ReadTimeOut) })
	routine.Go(func() { client.session.write() })
	return
}

func (client *TCPClient) heartbeat() {
	t := time.NewTicker(1 * time.Second)
	p := &message.ReqHeartbeat{}
	defer t.Stop()
loop:
	for {
		select {
		case <-client.heartClose:
			break loop

		case <-t.C:
			if !client.isRunning() {
				break loop
			}
			client.Send(message.MSGID_ReqHeartbeatE, p)
		}
	}
	logger.Infof("远程服务器心跳线程退出, Addr=%s", client.addr)
}

func (client *TCPClient) Reconnect() {
	if !client.isRunning() {
		return
	}

	client.connect()
}

func (client *TCPClient) Send(msgid message.MSGID, pbmsg proto.Message) error {
	if !client.isRunning() {
		return fmt.Errorf("client:%d is not running", client.cid)
	}
	sess := client.session
	if sess != nil {
		return sess.Send(msgid, pbmsg)
	}
	return fmt.Errorf("client:%d session is nil", client.cid)
}

func (client *TCPClient) RemoteAddr() net.Addr {
	return client.session.RemoteAddr()
}

func (client *TCPClient) Addr() string {
	return client.addr
}

func (client *TCPClient) Session() *core.Session {
	return client.session.userSession
}

type Role struct {
	Id int64
}

func (role *Role) ID() int64 {
	return role.Id
}
func (client *TCPClient) SetEntity(roleId int64) {
	role := &Role{Id: roleId}
	if client.session != nil {
		client.session.userSession.SetEntity(role)
	}

}
func (client *TCPClient) GetEntity() int64 {
	if client.session == nil || client.session.userSession == nil || client.session.userSession.Entity() == nil {
		return 0
	}
	return client.session.userSession.Entity().ID()
}
