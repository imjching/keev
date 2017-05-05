package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/imjching/go-kvs/kvs"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:1234"
)

func storeItem(client kvs.KVSClient, in *kvs.StoreRequest) {
	resp, err := client.StoreItem(context.Background(), in)
	if err != nil {
		log.Fatalf("Could not store item: %v", err)
	}
	if resp.Success {
		log.Printf("A new item has been stored!")
	}
}

func loadItem(client kvs.KVSClient, in *kvs.LoadRequest) {
	resp, err := client.LoadItem(context.Background(), in)
	if err != nil {
		log.Printf("Could not load item: %v", err)
		return
	}
	log.Printf("A new item has been loaded! %s:%s", resp.Key, resp.Value)
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := kvs.NewKVSClient(conn)

	fmt.Println("go-kvs 1.0")
	fmt.Printf("[%s %s/%s]\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Println("Type \"help\" for more information.")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">>> ")
		scanner.Scan()
		command := strings.Fields(scanner.Text())
		if len(command) == 0 {
			fmt.Println("invalid!")
			continue
		}
		switch command[0] {
		case "put":
			if len(command) != 3 {
				fmt.Println("Invalid input: put <key> <value>")
				continue
			}
			sr := &kvs.StoreRequest{
				Key:   command[1],
				Value: command[2],
			}
			storeItem(client, sr)
		case "get":
			if len(command) != 2 {
				fmt.Println("Invalid input: get <key>")
				continue
			}
			lr := &kvs.LoadRequest{
				Key: command[1],
			}
			loadItem(client, lr)
		default:
			fmt.Println("Invalid input!")
		}
	}
}
