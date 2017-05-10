package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/carmark/pseudo-terminal-go/terminal"
	pb "github.com/imjching/keev/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address = "localhost:1234"
)

var username = flag.String("username", "", "Username")
var password = flag.String("password", "", "Password")

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

func printHelpMessage() {
	fmt.Println(`Usage: COMMAND [command-specific-options]

    set [key] [value]    # sets a key-value pair if not present
    update [key] [value] # updates a key-value pair if present
    has [key]            # determines if key is present
    unset [key]          # remove key from store
    get [key]            # retrieve key from store
    count                # retrieve number of key-value pairs in store
    show keys            # show all keys in store
    show data            # show all key-value pairs in store
    show namespaces      # show all namespaces in store
    use [namespace]      # select a namespace
	`)
}

func handleCommand(client pb.KVSClient, term *terminal.Terminal, command []string) bool {
	if len(command) == 0 {
		printHelpMessage()
		// fmt.Println("ERROR:  available options: set, update, has, unset, get, count, show, use")
		return true
	}
	switch strings.ToLower(command[0]) {
	case ".exit":
		return false
	case "help":
		printHelpMessage()
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
			break
		}
		str := UseNamespace(client, command[1])
		if str != "" {
			term.SetPrompt(*username + "@" + str + " > ")
		}
	default:
		fmt.Println("ERROR:  syntax error at or near \"" + command[0] + "\"")
	}
	return true
}

func main() {
	flag.Parse()
	if *username == "" {
		log.Fatalf("Please supply a username using the --username flag")
	}

	creds, err := credentials.NewClientTLSFromFile("keys/cert.pem", "localhost")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&loginCreds{
		Username: *username,
		Password: *password,
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

	term.SetPrompt(*username + " > ")
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
