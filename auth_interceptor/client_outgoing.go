package main

import (
	"fmt"
	pb "github.com/pandaychen/grpc-wrapper-framework/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"time"
)

const (
	Address         = "127.0.0.1:50052"
	OpenTLS         = true
	timestampFormat = time.StampNano // "Jan _2 15:04:05.000"
)

func main() {
	var err error
	var opts []grpc.DialOption

	if OpenTLS {
		creds, err := credentials.NewClientTLSFromFile("server.pem", "xxxx")
		if err != nil {
			grpclog.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(Address, opts...)

	if err != nil {
		grpclog.Fatalln(err)
	}

	defer conn.Close()

	c := pb.NewGreeterServiceClient(conn)
	md := metadata.Pairs("timestamp", time.Now().Format(timestampFormat), "id", "10001", "ssokey", "abcdefg")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	reqBody := new(pb.HelloRequest)
	reqBody.Name = "pandaychen"

	//use ctx to pass metadata
	r, err := c.SayHello(ctx, reqBody)
	if err != nil {
		fmt.Println(err)
		grpclog.Fatalln(err)
	}

	grpclog.Println(r.Message)
}
