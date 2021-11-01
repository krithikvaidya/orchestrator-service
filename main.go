package main

import (
	"context"
	"fmt"
	"log"
	"net"

	constants "github.com/krithikvaidya/orchestrator-service/constants"
	user_pb "github.com/krithikvaidya/orchestrator-service/user"
	"google.golang.org/grpc"
)

type orchestrator1Server struct {
	user_pb.UnimplementedOrchestrator1ServiceServer
}

func (orc1Server *orchestrator1Server) GetUserByName(ctx context.Context, name *user_pb.UserName) (*user_pb.User, error) {
	return nil, fmt.Errorf("not implemented yet. %v will implement me", name.Name)
}

func main() {

	// Create TCP listener to bind gRPC server to
	listener, err := net.Listen("tcp", constants.ORC1_ADDR)
	if err != nil {
		log.Fatalf("Cannot bind server to tcp port %v,\nError: %v", constants.ORC1_ADDR, err)
	}

	orc1Server := grpc.NewServer()
	user_pb.RegisterOrchestrator1ServiceServer(orc1Server, &orchestrator1Server{})

	// Start the server
	log.Printf("\nStarting gRPC server at port %v...\n", constants.ORC1_ADDR)
	err = orc1Server.Serve(listener)

	if err != nil {
		log.Fatalf("Error occurred in gRPC server Serve(): %v", err)
	}
}
