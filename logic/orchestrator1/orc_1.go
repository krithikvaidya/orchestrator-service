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

type orchestrator1 struct {
	user_pb.UnimplementedOrchestrator1ServiceServer

	orc1Server *grpc.Server
	orc2Client user_pb.Orchestrator2ServiceClient
	shutdown   chan string
}

func (orc1 *orchestrator1) GetUserByName(ctx context.Context, name *user_pb.UserName) (*user_pb.User, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error

	// Call GetUser RPC
	response, err := orc1.orc2Client.GetUser(ctx, &user_pb.User{Name: name.Name})
	if err != nil {
		log.Printf("Got error in GetUserByName: %v", err)
	} else {
		log.Printf("Got response in GetUserByName: %v", response)
	}

	return response, err
}

func (orc1 *orchestrator1) startGRPCServer(ctx context.Context, listener net.Listener) {

	// Shut down the gRPC server if the context is cancelled
	go func() {

		// Block till the context is cancelled
		<-ctx.Done()

		// Stop the server from accepting new connections but
		// allow pending RPCs to complete.
		orc1.orc1Server.GracefulStop()
		orc1.shutdown <- "gRPC server shutdown successful."

	}()

	// Start the server
	log.Printf("\nStarting gRPC server for Orchestrator 1 at port %v...\n", constants.ORC1_ADDR)
	err := orc1.orc1Server.Serve(listener)

	if err != nil {
		log.Fatalf("Error occurred in Orchestrator 1 gRPC server Serve(): %v", err)
	}
}

// Listen for termination signal and ensure graceful shutdown.
func (orc1 *orchestrator1) listenForShutdown(cancel context.CancelFunc) {

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
	case str := <-orc1.shutdown:
		log.Printf("Shutdown: %v", str)
	case <-time.After(5 * time.Second):
		log.Printf("\nTimeout expired, force shutdown invoked.\n")
		return
	}

	log.Printf("Shutdown complete")
}

func main() {

	// Create TCP listener which the gRPC server will use to listen for requests
	listener, err := net.Listen("tcp", constants.ORC1_ADDR)
	if err != nil {
		log.Fatalf("Cannot bind server to tcp port %v,\nError: %v", constants.ORC1_ADDR, err)
	}

	orc1 := &orchestrator1{
		shutdown: make(chan string),
	}
	orc1Server := grpc.NewServer()
	orc1.orc1Server = orc1Server

	// Bind the given instance of the orchestrator 1 struct, for every gRPC call
	user_pb.RegisterOrchestrator1ServiceServer(orc1Server, orc1)

	conn, err := grpc.Dial(constants.ORC2_ADDR, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to orchestrator's gRPC server. Error: %v", err)
	}
	defer conn.Close()

	// Obtain the gRPC client stub for the service
	orc2Client := user_pb.NewOrchestrator2ServiceClient(conn)
	orc1.orc2Client = orc2Client

	ctx, cancel := context.WithCancel(context.Background())

	go orc1.startGRPCServer(ctx, listener)

	orc1.listenForShutdown(cancel)
}
