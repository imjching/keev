package main

import (
	"fmt"

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
	// md := metadata.Pairs("token", "Bearer XXXX", "token", "asdfo")
	// ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := client.Set(currentCtx(), &pb.KeyValuePair{Key: key, Value: value})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	fmt.Println(resp.Value)
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

func UseNamespace(client pb.KVSClient, namespace string) string {
	resp, err := client.UseNamespace(currentCtx(), &pb.Namespace{Namespace: namespace})
	if err != nil {
		fmt.Println("ERROR: ", err)
		return ""
	}

	token = resp.Token
	return namespace
}

// type KVSClient interface {
//     // Inserts a key-value pair into a namespace, if not present
//     Set(ctx context.Context, in *KeyValuePair, opts ...grpc.CallOption) (*Response, error)
//     // Updates a key-value pair in a namespace, if present
//     Update(ctx context.Context, in *KeyValuePair, opts ...grpc.CallOption) (*Response, error)
//     // Checks if a key is in a namespace
//     Has(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Response, error)
//     // Removes a key in a namespace, if present
//     Unset(ctx context.Context, in *Key, opts ...grpc.CallOption) (*KeyValuePair, error)
//     // Retrieves an element from a namespace under given key
//     Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*KeyValuePair, error)
//     // Returns the total number of key-value pairs in a namespace
//     Count(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*CountResponse, error)
//     // Retrieve all keys in a namespace
//     ShowKeys(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*ShowKeysResponse, error)
//     // Retrieve all key-value pairs in a namespace
//     ShowData(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*ShowDataResponse, error)
//     // Retrieve all namespaces in the key-value store that belongs to the user
//     // NOTE: No token needed
//     ShowNamespaces(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*ShowNamespacesResponse, error)
//     // Changes the current namespace, returns a token that must be used for
//     // subsequent requests
//     // NOTE: No token needed
//     UseNamespace(ctx context.Context, in *Namespace, opts ...grpc.CallOption) (*NamespaceResponse, error)
// }
