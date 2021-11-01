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

type mockUserDataServer struct {
	user_pb.UnimplementedMockUserDataServiceServer
}

func (mockDataServer *mockUserDataServer) GetMockUserData(ctx context.Context, name *user_pb.UserName) (*user_pb.User, error) {

	var err error = nil
	var user *user_pb.User

	if len(name.Name) < 6 {
		err = fmt.Errorf("invalid name %v. Name length must be atleast 6 characters", name.Name)
	} else {
		user = &user_pb.User{
			Name:  name.Name,
			Class: fmt.Sprint(len(name.Name)),
			Roll:  int64(len(name.Name)) * 10,
		}
	}

	return user, err
}

func main() {

	// Create TCP listener which the gRPC server will use to listen for requests
	listener, err := net.Listen("tcp", constants.DUMMY_DATA_SERV_ADDR)
	if err != nil {
		log.Fatalf("Cannot bind server to tcp port %v,\nError: %v", constants.DUMMY_DATA_SERV_ADDR, err)
	}

	mockDataServer := grpc.NewServer()
	user_pb.RegisterMockUserDataServiceServer(mockDataServer, &mockUserDataServer{})

	// Start the server
	log.Printf("\nStarting mock user data gRPC service at port %v...\n", constants.DUMMY_DATA_SERV_ADDR)
	err = mockDataServer.Serve(listener)

	if err != nil {
		log.Fatalf("Error occurred in mock user data gRPC server Serve(): %v", err)
	}
}
