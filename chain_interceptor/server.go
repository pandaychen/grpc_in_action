package main

import (
	"fmt"
	multiint "github.com/mercari/go-grpc-interceptor/multiinterceptor"
	pb "github.com/pandaychen/grpc-wrapper-framework/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"       // grpc 响应状态码
	"google.golang.org/grpc/credentials" // grpc认证包
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata" // grpc metadata包
	"net"
)

const (
	Address = "127.0.0.1:50052"
	ID      = "10001"
	SSOKEY  = "abcdefg"
)

func authUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("auth")
	err := auth(ctx)
	if err != nil {
		fmt.Println("auth error=", err)
		return nil, err
	}
	return handler(ctx, req)
}

func fooUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("foo")
	newctx := context.WithValue(ctx, "foo_key", "foo_value")
	rsp, err := handler(newctx, req)
	fmt.Printf("after foo,rsp=%v,err=%v\n", rsp, err)
	return rsp, err
}

func barUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("bar")
	newctx := context.WithValue(ctx, "bar_key", "bar_value")
	rsp, err := handler(newctx, req)
	fmt.Printf("after bar,rsp=%v,err=%v\n", rsp, err)
	return rsp, err
}

type helloService struct{}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println(ctx)
	fmt.Println(ctx.Value("foo_key"), ctx.Value("bar_key"))
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
	uIntOpt := grpc.UnaryInterceptor(multiint.NewMultiUnaryServerInterceptor(
		authUnaryInterceptor,
		fooUnaryInterceptor,
		barUnaryInterceptor,
	))

	opts = append(opts, uIntOpt)

	//grpc Server init
	s := grpc.NewServer(opts...)

	//register HelloService
	pb.RegisterGreeterServiceServer(s, HelloService)

	s.Serve(listen)
}
