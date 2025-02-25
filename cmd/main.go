package main

import (
	"context"
	"fmt"
	"net"

	pb "github.com/escape-ship/ordersrv/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedYourServiceServer
}

// Echo implements helloworld.GreeterServer
func (s *server) Echo(_ context.Context, in *pb.StringMessage) (*pb.StringMessage, error) {
	fmt.Println("Echo: " + in.Value)
	return &pb.StringMessage{Value: "Hello " + in.Value}, nil
}

func main() {
	fmt.Println("hello world")

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return
	}

	s := grpc.NewServer()

	pb.RegisterYourServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		return
	}
}
