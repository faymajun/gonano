package rpc

import (
	"github.com/faymajun/gonano/message"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var log = logrus.WithField("com", "rpc")

// HelloServer is used to implement helloworld.Server.
type HelloServer struct {
}

func (s *HelloServer) HelloWorld(c context.Context, request *message.HelloRequest) (*message.HelloResponse, error) {
	log.Printf("Received: %v", request.Request)
	return &message.HelloResponse{Response: "hello, " + request.Request}, nil
}

func (s *HelloServer) HelloWorldServerStream(request *message.HelloRequest, server message.HelloService_HelloWorldServerStreamServer) error {
	panic("implement me")
}

func (s *HelloServer) HelloWorldClientStream(server message.HelloService_HelloWorldClientStreamServer) error {
	panic("implement me")
}

func (s *HelloServer) HelloWorldClientAndServerStream(server message.HelloService_HelloWorldClientAndServerStreamServer) error {
	for {
		//接受客户端消息
		req, err := server.Recv()
		if err != nil && err.Error() == "EOF" {
			log.Info("HelloWorldClientAndServerStream 关闭")
			break
		}
		if err != nil {
			log.Errorf("HelloWorldClientAndServerStream err:%v", err)
			break
		} else {
			log.Printf("HelloWorldClientAndServerStream %v", req.Request)
			//返回客户端结果
			server.Send(&message.HelloResponse{Response: "hello my is gRpcServer stream"})
		}
	}
	return nil
}

var (
	serverPort       = ":2222" // 监听地址
	heartbeat  int64 = 5       // 心跳间隔
)
