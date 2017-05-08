package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

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

var (
	AccessDeniedErr = errors.New("access denied")

// 	EmptyMetadataErr = errors.New("empty metadata")
// 	EmptyTokenErr    = errors.New("empty token")
// 	InvalidToken     = errors.New("invalid token")
)

var mySigningKey = []byte("AllYourBase")

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
		fmt.Println(md)
		if len(md["username"]) > 0 && md["username"][0] == "admin" &&
			len(md["password"]) > 0 && md["password"][0] == "admin123" {
			return nil
		}

		return AccessDeniedErr
	}

	return EmptyMetadataErr
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

	file, err := os.Open("data/users.json")
	if err != nil {
		log.Fatalln("Unable to load users", err)
	}

	users := auth.NewCredentialsStore()
	if err := users.Load(file); err != nil {
		log.Fatalf("failed to load credentials: %s", err.Error())
	}

	fmt.Println(users)

	s := grpc.NewServer(
		grpc.Creds(cert),
		grpc.StreamInterceptor(streamInterceptor),
		grpc.UnaryInterceptor(unaryInterceptor),
	)

	protobuf.RegisterKVSServer(s, NewServer())

	log.Printf("Listening RPC Server on port localhost%s", port)
	s.Serve(listener)
}
