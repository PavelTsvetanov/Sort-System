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
	sortingRobotClient, conn := newSortingRobotClient()
	defer conn.Close()

	grpcServer, lis := newFulfilmentServer(sortingRobotClient)

	fmt.Printf("gRPC server started. Listening on %s\n", serverPort)
	grpcServer.Serve(lis)
}

func newSortingRobotClient() (gen.SortingRobotClient, *grpc.ClientConn) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(sortingServerAddress, opts...)
	if err != nil {
		log.Fatalf("Failed ot dial sorting robot server: %s", err)
	}
	return gen.NewSortingRobotClient(conn), conn
}

func newFulfilmentServer(sortingRobot gen.SortingRobotClient) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	gen.RegisterFulfillmentServer(grpcServer, newFulfilmentService(sortingRobot))
	reflection.Register(grpcServer)

	return grpcServer, lis
}
