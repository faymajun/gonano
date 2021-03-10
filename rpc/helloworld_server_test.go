package rpc

import (
	"github.com/faymajun/gonano/consul"
	"github.com/faymajun/gonano/message"
	"github.com/faymajun/gonano/util"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	// 监听端口，设置心跳，设置包最大值，设置每秒最多包数量，开启每个客户端一个处理协程
	consul.RegisterService("dc1", "helloworld", 1, "127.0.0.1", 222)
	lis, err := net.Listen("tcp", ":222")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	message.RegisterHelloServiceServer(server, &HelloServer{})
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	logrus.Warnf("服务器启动成功")
	go func() {
		// make sure Serve() is called
		time.Sleep(time.Second * 500)
		server.GracefulStop()
	}()

	err = server.Serve(lis)
	if err != nil {
		t.Fatalf("Serve() returned non-nil error on GracefulStop: %v", err)
	}
	logrus.Infof("Server shutdown Finish!!!")
}

func TestStartServer2(t *testing.T) {
	// 监听端口，设置心跳，设置包最大值，设置每秒最多包数量，开启每个客户端一个处理协程
	consul.RegisterService("dc1", "helloworld", 2, util.LocalIPString(), 333)
	lis, err := net.Listen("tcp", ":333")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	message.RegisterHelloServiceServer(server, &HelloServer{})
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	logrus.Warnf("服务器启动成功")
	go func() {
		// make sure Serve() is called
		time.Sleep(time.Second * 5000)
		server.GracefulStop()
	}()

	err = server.Serve(lis)
	if err != nil {
		t.Fatalf("Serve() returned non-nil error on GracefulStop: %v", err)
	}
	server.GracefulStop()
	logrus.Infof("Server shutdown Finish!!!")
}
