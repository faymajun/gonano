package rpc

import (
	"github.com/faymajun/gonano/message"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"testing"
	"time"
)

func TestStartClient(t *testing.T) {
	// Set up a connection to the server.
	//ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	resolver.Register(NewBuilder())
	//conn, err := rpc.DialContext(ctx, "http://127.0.0.1:2000/v1/health/service/game-gateway-server?dc=dc1", rpc.WithBlock(), rpc.WithInsecure(), rpc.WithBalancerName("round_robin"))
	//client, err := rpc.Dial("consul://192.168.10.15:8500/dc1")
	//client, err := rpc.Dial("consul://http://192.168.10.15:8500/gateway")
	client, err := grpc.Dial("consul://127.0.0.1:2000/helloworld", grpc.WithInsecure(), grpc.WithBalancerName("round_robin")) // 地址：http://127.0.0.1:2000/v1/health/service/helloworld?passing=1
	//client, err := rpc.Dial(
	//	"consul://127.0.0.1:2000/gateway",
	//	rpc.WithInsecure(),
	//	rpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	//)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer client.Close()
	c := message.NewHelloServiceClient(client)

	for {
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		r, err := c.HelloWorld(ctx, &message.HelloRequest{Request: "rpc client!"})
		if err != nil {
			log.Fatalf("could not rpc: %v", err)
		}
		log.Printf("rpc test: %s", r.Response)
		time.Sleep(time.Second * 2)
	}
}

// consul://127.0.0.1:2000/helloworld转换http地址：http://127.0.0.1:2000/v1/health/service/helloworld?passing=1
// passing=1 表示service check passing  服务健康检测通过
func TestStartClient2(t *testing.T) {
	resolver.Register(NewBuilder())
	client, err := grpc.Dial("consul://127.0.0.1:2000/helloworld", grpc.WithInsecure(), grpc.WithBalancerName("round_robin"))
	//client, err := rpc.Dial("consul://127.0.0.1:2000/cloud-rpc-server-consul", rpc.WithInsecure(), rpc.WithBalancerName("round_robin"))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer client.Close()
	c := message.NewHelloServiceClient(client)

	//ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	r, err := c.HelloWorldClientAndServerStream(context.Background(), grpc.EmptyCallOption{})
	if err != nil {
		log.Fatalf("could not rpc: %v", err)
	}
	for i := 0; i < 100; i++ {
		err := r.Send(&message.HelloRequest{Request: "my is golang gRpc client"})
		if err != nil {
			log.Error(err)
			c = message.NewHelloServiceClient(client)
			r, err = c.HelloWorldClientAndServerStream(context.Background(), grpc.EmptyCallOption{})
			if err != nil {
				log.Fatalf("could not rpc: %v", err)
			}
		}
		time.Sleep(time.Second * 1)
	}
	// 发送完毕
	r.CloseSend()
	//循环接受服务端返回的结果，直到返回EOF
	for {
		res, err := r.Recv()
		if err != nil && err.Error() == "EOF" {
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
			break
		}
		log.Printf("result:%v", res.Response)
	}
}
