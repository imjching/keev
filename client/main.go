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
	// "google.golang.org/grpc/metadata"
)

const (
	address = "localhost:1234"
)

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

func handleCommand(client pb.KVSClient, term *terminal.Terminal, command []string) bool {
	if len(command) == 0 {
		fmt.Println("ERROR:  available options: set, update, has, unset, get, count, show, use")
		return true
	}
	switch strings.ToLower(command[0]) {
	case ".exit":
		return false
	case "set":
		if len(command) != 3 {
			fmt.Println("ERROR:  syntax error. use \"put [key] [value]\"")
			break
		}
		Set(client, command[1], command[2])
	case "update":
		if len(command) != 3 {
			fmt.Println("ERROR:  syntax error. use \"update [key] [value]\"")
			break
		}
		Update(client, command[1], command[2])
	case "has":
		if len(command) != 2 {
			// TODO: implement smart guessing?
			fmt.Println("ERROR:  syntax error. use \"has [key]\"")
			break
		}
		Has(client, command[1])
	case "unset":
		if len(command) != 2 {
			// TODO: implement smart guessing?
			fmt.Println("ERROR:  syntax error. use \"unset [key]\"")
			break
		}
		Unset(client, command[1])
	case "get":
		if len(command) != 2 {
			// TODO: implement smart guessing?
			fmt.Println("ERROR:  syntax error. use \"get [key]\"")
			// fmt.Println("ERROR:  key \"" + command[1] + "\" does not exist")
			break
		}
		Get(client, command[1])
	case "count":
		Count(client)
	case "show":
		if len(command) != 2 {
			fmt.Println("ERROR:  syntax error. use \"show [keys|data|namespaces]\"")
			break
		}
		Show(client, command[1])
	case "use":
		if len(command) != 2 {
			fmt.Println("ERROR:  syntax error. use \"use [namespace]\"")
			// fmt.Println("ERROR:  namespace \"" + command[1] + "\" does not exist")
			break
		}
		str := UseNamespace(client, command[1])
		if str != "" {
			term.SetPrompt("imjching@" + str + " > ")
		}
	default:
		fmt.Println("ERROR:  syntax error at or near \"" + command[0] + "\"")
	}
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
		if !handleCommand(client, term, command) {
			return
		}
		line, err = term.ReadLine()

	}
	term.Write([]byte(line))
}
