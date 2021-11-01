package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	constants "github.com/krithikvaidya/orchestrator-service/constants"
	user_pb "github.com/krithikvaidya/orchestrator-service/user"
	"google.golang.org/grpc"
)

type orchestrator2 struct {
	user_pb.UnimplementedOrchestrator2ServiceServer

	orc2Server *grpc.Server
	userClient user_pb.MockUserDataServiceClient
	shutdown   chan string
}

func (orc2 *orchestrator2) GetUser(ctx context.Context, user *user_pb.User) (*user_pb.User, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error

	// Call GetMockUserData RPC
	response, err := orc2.userClient.GetMockUserData(ctx, &user_pb.UserName{Name: user.Name})
	if err != nil {
		log.Printf("Got error in GetUser: %v", err)
	} else {
		log.Printf("Got response in GetUser: %v", response)
	}

	return response, err
}

func (orc2 *orchestrator2) startGRPCServer(ctx context.Context, listener net.Listener) {

	// Shut down the gRPC server if the context is cancelled
	go func() {

		// Block till the context is cancelled
		<-ctx.Done()

		// Stop the server from accepting new connections but
		// allow pending RPCs to complete.
		orc2.orc2Server.GracefulStop()
		orc2.shutdown <- "gRPC server shutdown successful."

	}()

	// Start the server
	log.Printf("\nStarting gRPC server for Orchestrator 2 at port %v...\n", constants.ORC2_ADDR)
	err := orc2.orc2Server.Serve(listener)

	if err != nil {
		log.Fatalf("Error occurred in Orchestrator 2 gRPC server Serve(): %v", err)
	}
}

// Listen for termination signal and ensure graceful shutdown.
func (orc2 *orchestrator2) listenForShutdown(cancel context.CancelFunc) {

	// Capture termination signals
	osSigs := make(chan os.Signal, 1)                      // Listen for OS signals, with buffer size 1
	signal.Notify(osSigs, syscall.SIGTERM, syscall.SIGINT) // SIGKILL and SIGSTOP cannot be caught by a program

	rcvdSig := <-osSigs

	log.Printf("\n\nTermination signal received: %v\n", rcvdSig)

	signal.Stop(osSigs) // Stop listening for signals
	close(osSigs)

	cancel() // Here this will only signal ctx.Done() in StartGRPCServer's goroutine

	// Listen for the goroutines (here only 1 goroutine, StartGRPCServer) to finish
	// their shutdown and write to the shutdown channel.
	select {
	case str := <-orc2.shutdown:
		log.Printf("Shutdown: %v", str)
	case <-time.After(5 * time.Second):
		log.Printf("\nTimeout expired, force shutdown invoked.\n")
		return
	}

	log.Printf("Shutdown complete")
}

func main() {

	// Create TCP listener which the gRPC server will use to listen for requests
	listener, err := net.Listen("tcp", constants.ORC2_ADDR)
	if err != nil {
		log.Fatalf("Cannot bind server to tcp port %v,\nError: %v", constants.ORC2_ADDR, err)
	}

	orc2 := &orchestrator2{
		shutdown: make(chan string),
	}
	orc2Server := grpc.NewServer()
	orc2.orc2Server = orc2Server

	user_pb.RegisterOrchestrator2ServiceServer(orc2Server, orc2)

	conn, err := grpc.Dial(constants.DUMMY_DATA_SERV_ADDR, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to orchestrator's gRPC server. Error: %v", err)
	}
	defer conn.Close()

	// Obtain the gRPC client stub for the service
	orc2Client := user_pb.NewMockUserDataServiceClient(conn)
	orc2.userClient = orc2Client

	ctx, cancel := context.WithCancel(context.Background())

	go orc2.startGRPCServer(ctx, listener)

	orc2.listenForShutdown(cancel)
}
