package main

import (
	"fmt"
	"log"
	"net"

	"github.com/PavelTsvetanov/sort-system/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serverPort = "localhost:10001"

func main() {
	grpcServer, lis := newFulfilmentServer()

	fmt.Printf("gRPC server started. Listening on %s\n", serverPort)
	grpcServer.Serve(lis)
}

func newFulfilmentServer() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	gen.RegisterFulfillmentServer(grpcServer, newFulfilmentService())
	reflection.Register(grpcServer)

	return grpcServer, lis
}
