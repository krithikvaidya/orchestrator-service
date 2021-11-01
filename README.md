# Orchestrator Service

## Part 2

- The root of the repo contains `main.go`, which implements the server (returns a hardcoded value for now)
- The *client* folder contains the client code which connects to the gRPC server using the client stub, and invokes the *GetUserByName* RPC.
- The *constants* folder contains some constants used by our program
- The *user* folder contains the `user.proto` file (protobufs and RPC definitions), along with the generated *.pb.go* files

## Output

- Server:
![](screenshots/part2_server.png)

- Client
![](screenshots/part2_client.png)