package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Just-Goo/my-grpc-proto/protogen/go/hello"
)

// unary server
func (g *GrpcAdapter) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	greet := g.helloService.GenerateHello(req.Name)

	return &hello.HelloResponse{
		Greet: greet,
	}, nil
}

// server streaming
func (g *GrpcAdapter) SayManyHellos(req *hello.HelloRequest, stream hello.HelloService_SayManyHellosServer) error {

	for i := 0; i < 10; i++ {
		greet := g.helloService.GenerateHello(req.Name)

		res := fmt.Sprintf("[%d] %s", i, greet)

		stream.Send(
			&hello.HelloResponse{
				Greet: res,
			},
		)

		time.Sleep(500 * time.Millisecond) // pause for half a second after each response
	}

	return nil
}

// client streaming
func (g *GrpcAdapter) SayHelloToEveryone(stream hello.HelloService_SayHelloToEveryoneServer) error {

	res := ""

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(
				&hello.HelloResponse{
					Greet: res,
				},
			)
		}

		if err != nil {
			log.Fatalln("Error while reading from client:", err)
		}

		greet := g.helloService.GenerateHello(req.Name)

		res += greet + " => "
	}
}

// bi-directional streaming
func (g *GrpcAdapter) SayHelloContinuous(stream hello.HelloService_SayHelloContinuousServer) error {
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalln("Error while reading from client:", err)
		}

		greet := g.helloService.GenerateHello(req.Name)

		// send response to client immediately you receive a request
		err = stream.Send(
			&hello.HelloResponse{
				Greet: greet,
			},
		)

		if err != nil {
			log.Fatalln("Error while sending response to client:", err)
		}
	}
}
