package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/imjching/keev/auth"
	pb "github.com/imjching/keev/protobuf"
	"github.com/orcaman/concurrent-map"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	port = ":1234"
)

var (
	AccessDeniedErr  = errors.New("access denied")
	EmptyMetadataErr = errors.New("empty metadata")
	EmptyTokenErr    = errors.New("empty token")
	InvalidToken     = errors.New("invalid token")
)

var mySigningKey = []byte("AllYourBase")

type server struct {
	data cmap.ConcurrentMap
}

type NamespaceClaims struct {
	Username string `json:"user"`
	Database string `json:"foo"`
	jwt.StandardClaims
}

func (s *server) ChangeNamespace(ctx context.Context, in *pb.StoreRequest) (*pb.StoreResponse, error) {
	// Verify if user has permission to the namespace
	// TODO

	claims := NamespaceClaims{
		"admin",
		"project1",
		jwt.StandardClaims{
			Issuer: "HashDB",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	fmt.Printf("%v %v", ss, err)

	return &pb.StoreResponse{Success: true}, nil
}

func (s *server) StoreItem(ctx context.Context, in *pb.StoreRequest) (*pb.StoreResponse, error) {
	log.Printf("[STORE] %s:%s", in.Key, in.Value)

	s.data.Set(in.Key, in.Value)
	log.Printf("[STORE SUCCESSFUL] %s:%s", in.Key, in.Value)

	fmt.Println(s.data.Items())

	return &pb.StoreResponse{Success: true}, nil
}

func (s *server) LoadItem(ctx context.Context, in *pb.LoadRequest) (*pb.LoadResponse, error) {
	log.Printf("[LOAD] %s", in.Key)

	value, ok := s.data.Get(in.Key)
	if !ok {
		log.Printf("[LOAD UNSUCCESSFUL] Invalid key: %s", in.Key)
		return nil, errors.New(fmt.Sprintf("Invalid key: %s", in.Key))
	}

	value, ok = value.(string)

	log.Printf("[LOAD SUCCESSFUL] %s:%s", in.Key, value)
	return &pb.LoadResponse{Key: in.Key, Value: value.(string)}, nil
}

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
	fmt.Println(info.FullMethod)
	// methods which require a token to identify namespace
	switch info.FullMethod {
	case "/protobuf.KVS/StoreItem":
		// parse the database
		//token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		//     return []byte("AllYourBase"), nil
		// })

		// if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		//     fmt.Printf("%v %v", claims.Foo, claims.StandardClaims.ExpiresAt)
		// } else {
		//     fmt.Println(err)
		// }

		if md, ok := metadata.FromContext(ctx); ok {
			if len(md["token"]) == 0 {
				return nil, InvalidToken
			}
			// verify token here
		} else {
			return nil, EmptyTokenErr
		}
	}

	ctx = context.WithValue(ctx, "token", "asdfff")

	return handler(ctx, req)
}

func authorize(ctx context.Context) error {
	if md, ok := metadata.FromContext(ctx); ok {
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
	pb.RegisterKVSServer(s, &server{data: cmap.New()})

	log.Printf("Listening RPC Server on port localhost%s", port)
	s.Serve(listener)
}
