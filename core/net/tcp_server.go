package net

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/core/routine"
)

var serverId uint32
var servers sync.Map

func StopTcpServer() {
	logger.Infof("<<<servers is stop start>>>")
	servers.Range(func(k, v interface{}) bool {
		if s, ok := v.(*TCPServer); ok {
			s.Close()
		}
		return true
	})
	logger.Infof("<<<servers is stop over>>>")
}

type TCPServer struct {
	sid          uint32       // 唯一标识userSession
	listener     net.Listener // 监听器
	state        int32        // 状态
	addr         string       // 监听地址
	heartbeat    int64        // 心跳超时时间
	codecFactory func() Codec
}

func (server *TCPServer) Close() error {
	if !atomic.CompareAndSwapInt32(&server.state, TCP_AVAI, TCP_STOP) {
		return ErrDupClosed
	}
	servers.Delete(server.sid)
	return server.listener.Close()
}

func StartTcpServer(addr string, handler SessionHandler, codecFactory func() Codec, heartbeat int64, maxRecvSize uint32, maxRecvNum uint16, isProcess bool) error {
	return StartTcpProcessServer(addr, handler, codecFactory, heartbeat, maxRecvSize, maxRecvNum, isProcess, WriteQueLen, ReadQueLen)
}

func StartTcpProcessServer(addr string, handler SessionHandler, codecFactory func() Codec, heartbeat int64, maxRecvSize uint32, maxRecvNum uint16, isProcess bool, wQueLen, rQueLen int) error {
	if handler == nil {
		return fmt.Errorf("SessionHandler can not be nil")
	}

	lc := net.ListenConfig{KeepAlive: -1}
	listen, err := lc.Listen(context.Background(), "tcp4", addr)
	if err != nil {
		return fmt.Errorf("tcp listen on %s failed, errstr:%s", addr, err.Error())
	}
	ts := &TCPServer{
		sid:          atomic.AddUint32(&serverId, 1),
		listener:     listen,
		state:        TCP_AVAI,
		heartbeat:    heartbeat,
		codecFactory: codecFactory,
	}
	servers.Store(ts.sid, ts)

	// check availability
	if wQueLen <= 0 {
		wQueLen = WriteQueLen
	}
	if rQueLen <= 0 {
		rQueLen = ReadQueLen
	}

	routine.Go(func() { ts.start(handler, maxRecvSize, maxRecvNum, isProcess, wQueLen, rQueLen) })
	return nil
}

func (server *TCPServer) start(handler SessionHandler, maxRecvSize uint32, maxRecvNum uint16, isProcess bool, wQueLen, rQueLen int) {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			logger.Warnf("tcp accept failed:%s", err.Error())
			break
		}

		tcpSession := newTcpSession(conn, server.codecFactory(), handler, maxRecvSize, maxRecvNum, isProcess, wQueLen, rQueLen)
		routine.Go(func() { tcpSession.read(time.Duration(server.heartbeat)) })
		routine.Go(func() { tcpSession.write() })
	}
	server.Close()
}
