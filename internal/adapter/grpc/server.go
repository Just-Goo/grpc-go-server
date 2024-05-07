package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/Just-Goo/grpc-go-server/internal/port"
	"github.com/Just-Goo/my-grpc-proto/protogen/go/bank"
	"github.com/Just-Goo/my-grpc-proto/protogen/go/hello"
	"google.golang.org/grpc"
)

type GrpcAdapter struct {
	helloService port.HelloServicePort
	bankService  port.BankServicePort
	grpcPort     int
	server       *grpc.Server
	hello.HelloServiceServer
	bank.BankServiceServer
}

func NewGrpcAdapter(helloService port.HelloServicePort, bankService port.BankServicePort, grpcPort int) *GrpcAdapter {
	return &GrpcAdapter{
		helloService: helloService,
		bankService:  bankService,
		grpcPort:     grpcPort,
	}
}

func (g *GrpcAdapter) Run() {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", g.grpcPort))
	if err != nil {
		log.Fatalf("failed to listen on port %d : %v\n", g.grpcPort, err)
	}

	log.Printf("server listening on port %d\n", g.grpcPort)

	grpcServer := grpc.NewServer()
	g.server = grpcServer

	hello.RegisterHelloServiceServer(grpcServer, g) // register the hello service server
	bank.RegisterBankServiceServer(grpcServer, g) // register the bank service server

	if err = grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve on port %d : %v\n", g.grpcPort, err)
	}
}

func (g *GrpcAdapter) Stop() {
	g.server.Stop()
}
