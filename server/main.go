package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/imjching/go-kvs/kvs"
	"github.com/orcaman/concurrent-map"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":1234"
)

type server struct {
	data cmap.ConcurrentMap
}

func (s *server) StoreItem(ctx context.Context, in *kvs.StoreRequest) (*kvs.StoreResponse, error) {
	log.Printf("[STORE] %s:%s", in.Key, in.Value)

	s.data.Set(in.Key, in.Value)
	log.Printf("[STORE SUCCESSFUL] %s:%s", in.Key, in.Value)

	fmt.Println(s.data.Items())

	return &kvs.StoreResponse{Success: true}, nil
}

func (s *server) LoadItem(ctx context.Context, in *kvs.LoadRequest) (*kvs.LoadResponse, error) {
	log.Printf("[LOAD] %s", in.Key)

	value, ok := s.data.Get(in.Key)
	if !ok {
		log.Printf("[LOAD UNSUCCESSFUL] Invalid key: %s", in.Key)
		return nil, errors.New(fmt.Sprintf("Invalid key: %s", in.Key))
	}

	value, ok = value.(string)

	log.Printf("[LOAD SUCCESSFUL] %s:%s", in.Key, value)
	return &kvs.LoadResponse{Key: in.Key, Value: value.(string)}, nil
}

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	kvs.RegisterKVSServer(s, &server{data: cmap.New()})

	log.Printf("Listening RPC Server on port localhost%s", port)
	s.Serve(listener)
}
