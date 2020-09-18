package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"       // grpc 响应状态码
	"google.golang.org/grpc/credentials" // grpc认证包
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata" // grpc metadata包
	"net"

	pb "github.com/pandaychen/grpc-wrapper-framework/proto"
)

const (
	Address = "127.0.0.1:50052"
	ID      = "10001"
	SSOKEY  = "abcdefg"
)

type helloService struct{}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println(in)
	resp := new(pb.HelloReply)
	resp.Message = "Hello " + in.Name + "."

	return resp, nil
}

func auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "none token msg")
	}

	var id string
	var key string

	if val, ok := md["id"]; ok {
		id = val[0]
	}

	if val, ok := md["ssokey"]; ok {
		key = val[0]
	}

	//call third-party auth
	if id != ID || key != SSOKEY {
		return grpc.Errorf(codes.Unauthenticated, "Token valid: id=%s, key=%s", id, key)
	}
	fmt.Println("check succ")
	return nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	creds, err := credentials.NewServerTLSFromFile("server.pem", "server.key")
	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
	}

	opts = append(opts, grpc.Creds(creds))

	// register interceptor
	var interceptor grpc.UnaryServerInterceptor
	interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		err = auth(ctx)
		if err != nil {
			return
		}
		// do real rpc request
		return handler(ctx, req)
	}
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	//grpc Server init
	s := grpc.NewServer(opts...)

	//register HelloService
	pb.RegisterGreeterServiceServer(s, HelloService)

	s.Serve(listen)
}
