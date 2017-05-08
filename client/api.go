package main

import (
	"fmt"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/imjching/keev/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var token string = ""

func currentCtx() context.Context {
	if token == "" {
		return context.Background()
	}
	md := metadata.Pairs("token", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return ctx
}

// Inserts a key-value pair into a namespace, if not present
func Set(client pb.KVSClient, key, value string) {
	resp, err := client.Set(currentCtx(), &pb.KeyValuePair{Key: key, Value: value})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Println(resp.Value)
}

// Updates a key-value pair in a namespace, if present
func Update(client pb.KVSClient, key, value string) {
	resp, err := client.Update(currentCtx(), &pb.KeyValuePair{Key: key, Value: value})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Println(resp.Value)
}

// Checks if a key is in a namespace
func Has(client pb.KVSClient, key string) {
	resp, err := client.Has(currentCtx(), &pb.Key{Key: key})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Println(resp.Value)
}

// Removes a key in a namespace, if present
func Unset(client pb.KVSClient, key string) {
	resp, err := client.Unset(currentCtx(), &pb.Key{Key: key})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Println("Removed entry: Key:", resp.Key, "Value:", resp.Value)
}

// Retrieves an element from a namespace under given key
func Get(client pb.KVSClient, key string) {
	resp, err := client.Get(currentCtx(), &pb.Key{Key: key})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Println("Key:", resp.Key, ", Value:", resp.Value)
}

// Returns the total number of key-value pairs in a namespace
func Count(client pb.KVSClient) {
	resp, err := client.Count(currentCtx(), &google_protobuf.Empty{})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Printf("Found %d key-value pair(s)\r\n", resp.Count)
}

func Show(client pb.KVSClient, key string) {
	switch key {
	case "keys": // Retrieve all keys in a namespace
		resp, err := client.ShowKeys(currentCtx(), &google_protobuf.Empty{})
		if err != nil {
			fmt.Println("ERROR: ", err)
			return
		}
		fmt.Println("Keys:", resp.Keys)
	case "data": // Retrieve all key-value pairs in a namespace
		resp, err := client.ShowData(currentCtx(), &google_protobuf.Empty{})
		if err != nil {
			fmt.Println("ERROR: ", err)
			return
		}
		fmt.Println("Data:", resp.Data)
	case "namespaces": // Retrieve all namespaces in the key-value store that belongs to the user
		resp, err := client.ShowNamespaces(currentCtx(), &google_protobuf.Empty{})
		if err != nil {
			fmt.Println("ERROR: ", err)
			return
		}
		fmt.Println("Namespaces:", resp.Namespaces)
	default:
		fmt.Println("ERROR:  syntax error. use \"show [keys|data|namespaces]\"")
	}
}

// Changes the current namespace, returns a token that must be used for subsequent requests
// NOTE: No token needed
func UseNamespace(client pb.KVSClient, namespace string) string {
	resp, err := client.UseNamespace(currentCtx(), &pb.Namespace{Namespace: namespace})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return ""
	}

	token = resp.Token
	return namespace
}
