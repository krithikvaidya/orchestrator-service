package main

import (
	"context"
	"log"

	constants "github.com/krithikvaidya/orchestrator-service/constants"
	user_pb "github.com/krithikvaidya/orchestrator-service/user"
	"google.golang.org/grpc"
)

func main() {

	/**
	Connect to the orchestrator's gRPC server
	WithBlock():    tells Dial not to return until the connection is established
				    or an error is encountered (makes it synchronous)
	WithInsecure(): disables requirement of transport layer security
	*/
	conn, err := grpc.Dial(constants.ORC1_ADDR, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to orchestrator's gRPC server. Error: %v", err)
	}
	defer conn.Close()

	// Obtain the gRPC client stub for the service
	c := user_pb.NewOrchestrator1ServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call GetUserByName RPC
	response, err := c.GetUserByName(ctx, &user_pb.UserName{Name: "Krithik"})
	if err != nil {
		log.Fatalf("Got error:\n%v", err)
	}
	log.Printf("Got response:\n%v", response)
}
