package rpc

import (
	"fmt"
	"github.com/faymajun/gonano/util"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

func StartTcpserver(port int) *grpc.Server {
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	logrus.Warnf("rpc server start success, ip:%s, port:%d", util.LocalIPString(), port)
	err = server.Serve(lis)
	if err != nil {
		logrus.Fatalf("Serve() returned non-nil error on GracefulStop: %v", err)
	}

	return server
}
