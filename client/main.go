package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/carmark/pseudo-terminal-go/terminal"
	pb "github.com/imjching/keev/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	address = "localhost:1234"
)

func storeItem(client pb.KVSClient, in *pb.StoreRequest) {

	md := metadata.Pairs("authorization", "Bearer XXXX")
	ctx := metadata.NewContext(context.Background(), md)

	resp, err := client.StoreItem(ctx, in)
	if err != nil {
		log.Fatalf("Could not store item: %v", err)
	}
	if resp.Success {
		log.Printf("A new item has been stored!")
	}
}

func loadItem(client pb.KVSClient, in *pb.LoadRequest) {
	resp, err := client.LoadItem(context.Background(), in)
	if err != nil {
		log.Printf("Could not load item: %v", err)
		return
	}
	log.Printf("A new item has been loaded! %s:%s", resp.Key, resp.Value)
}

type loginCreds struct {
	Username, Password string
}

func (c *loginCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return true
}

func main() {

	creds, err := credentials.NewClientTLSFromFile("keys/cert.pem", "localhost")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&loginCreds{
		Username: "admin",
		Password: "admin123",
	}))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewKVSClient(conn)

	term, err := terminal.NewWithStdInOut()
	if err != nil {
		panic(err)
	}
	defer term.ReleaseFromStdInOut() // defer this

	fmt.Println("keev (1.0)")
	fmt.Println("Type \"help\" for help.")
	fmt.Println()

	term.SetPrompt("imjching > ")
	line, err := term.ReadLine()
	for {
		if err == io.EOF {
			term.Write([]byte(line))
			fmt.Println()
			return
		}
		if (err != nil && strings.Contains(err.Error(), "control-c break")) || len(line) == 0 {
			line, err = term.ReadLine()
			continue
		}

		command := strings.Fields(line)
		if len(command) == 0 {
			fmt.Println("invalid!")
		} else {
			switch command[0] {
			case ".exit":
				return
			case "put":
				if len(command) != 3 {
					fmt.Println("Invalid input: put <key> <value>")
				} else {
					sr := &pb.StoreRequest{
						Key:   command[1],
						Value: command[2],
					}
					storeItem(client, sr)
				}
			case "get":
				if len(command) != 2 {
					fmt.Println("Invalid input: get <key>")
				} else {
					lr := &pb.LoadRequest{
						Key: command[1],
					}
					loadItem(client, lr)
				}
				term.SetPrompt("imjching@namespace > ")
			default:
				fmt.Println("Invalid input!")
			}
		}

		line, err = term.ReadLine()

	}
	term.Write([]byte(line))
}
