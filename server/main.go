package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imjching/keev/auth"
	"github.com/imjching/keev/protobuf"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	port = ":1234"
)

var users *auth.CredentialsStore

// middleware
func streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := authorize(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["username"]) == 0 || len(md["password"]) == 0 || !users.Check(md["username"][0], md["password"][0]) {
			return AccessDeniedErr // should close client's socket instead...
		}
		return nil
	}
	return EmptyMetadataErr
}

// for graceful shutdown
func saveToDisk(server *Server, forced bool) {
	b, err := json.Marshal(server)
	err = ioutil.WriteFile("./data/data.json", b, 0644)
	if err != nil {
		if forced {
			fmt.Println("Data loss...")
		} else {
			fmt.Println("Failed to write to file...Trying again...")
			saveToDisk(server, true)
		}
	}
	log.Println("Saved to disk")
}

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Load our TLS key pair to use for authentication
	cert, err := credentials.NewServerTLSFromFile("keys/cert.pem", "keys/key.pem")
	if err != nil {
		log.Fatalln("Unable to load cert", err)
	}

	// load and print users
	file, err := os.Open("data/users.json")
	if err != nil {
		log.Fatalln("Unable to load users", err)
	}
	users = auth.NewCredentialsStore()
	if err := users.Load(file); err != nil {
		log.Fatalf("failed to load credentials: %s", err.Error())
	}
	fmt.Println("[USERS]:", users)

	// register grpc server
	s := grpc.NewServer(
		grpc.Creds(cert),
		grpc.StreamInterceptor(streamInterceptor),
		grpc.UnaryInterceptor(unaryInterceptor),
	)
	server := NewServer()
	protobuf.RegisterKVSServer(s, server)

	// load data
	data, err := ioutil.ReadFile("./data/data.json")
	if err != nil {
		fmt.Println("No previous data found. Creating a new one...")
	}
	if x := json.Unmarshal(data, server); x != nil {
		fmt.Println("No previous data found. Creating a new one...")
	}

	// save to disk every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	quit := make(chan struct{})
	go func(s *Server) {
		for {
			select {
			case <-ticker.C:
				saveToDisk(s, false)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(server)

	// graceful shutdown
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func(q chan struct{}, s *Server) {
		<-c
		fmt.Println()
		close(q)
		saveToDisk(s, false)
		os.Exit(1)
	}(quit, server)

	// listen
	log.Printf("Listening RPC Server on port localhost%s", port)
	s.Serve(listener)
}
