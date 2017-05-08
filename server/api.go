package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/imjching/keev/cmap"
	"github.com/imjching/keev/common"
	pb "github.com/imjching/keev/protobuf"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var (
	KVPExistsErr        = errors.New("key already exists")
	KVPMissingErr       = errors.New("key does not exist")
	MissingTokenErr     = errors.New("missing token for namespace, use Use() to set a namespace")
	InvalidTokenErr     = errors.New("invalid token for namespace, use Use() to set a namespace")
	EmptyMetadataErr    = errors.New("missing metadata, please login again")
	TokenSigningErr     = errors.New("unable to sign token")
	InvalidNamespaceErr = errors.New("invalid namespace, alphanumerics only")
)

type Token struct {
	Username  string `json:"username"`
	Namespace string `json:"database"`
	jwt.StandardClaims
}

type Server struct {
	Data cmap.ConcurrentMap `json:"data"`
}

func NewServer() *Server {
	return &Server{
		Data: cmap.New(),
	}
}

func verifyToken(ctx context.Context) (*Token, error) {
	// check if metadata exists
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, MissingTokenErr
	}
	tokenString := md["token"]
	if len(tokenString) == 0 {
		return nil, MissingTokenErr
	}
	// retrieve the jwt parser
	token, err := jwt.ParseWithClaims(tokenString[0], &Token{}, func(token *jwt.Token) (interface{}, error) {
		return common.JWTSigningToken, nil
	})
	if err != nil {
		return nil, InvalidTokenErr
	}
	// check if token is valid
	if claims, ok := token.Claims.(*Token); ok && token.Valid {
		return claims, nil
	}
	// otherwise, token is invalid
	return nil, InvalidTokenErr
}

// Inserts a key-value pair into a namespace, if not present
func (s *Server) Set(ctx context.Context, in *pb.KeyValuePair) (*pb.Response, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}
	newKey := token.Username + "." + token.Namespace + "." + in.Key
	if !s.Data.SetIfAbsent(newKey, in.Value) {
		return nil, KVPExistsErr
	}
	return &pb.Response{Success: true, Value: "(1 pair(s) affected)"}, nil
}

// Updates a key-value pair in a namespace, if present
func (s *Server) Update(ctx context.Context, in *pb.KeyValuePair) (*pb.Response, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}
	newKey := token.Username + "." + token.Namespace + "." + in.Key
	if !s.Data.Has(newKey) {
		return nil, KVPMissingErr
	}
	s.Data.Set(newKey, in.Value)
	return &pb.Response{Success: true, Value: "(1 pair(s) affected)"}, nil
}

// Checks if a key is in a namespace
func (s *Server) Has(ctx context.Context, in *pb.Key) (*pb.Response, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}
	newKey := token.Username + "." + token.Namespace + "." + in.Key
	ok := s.Data.Has(newKey)
	if !ok {
		return &pb.Response{Success: false, Value: "(0 pair(s) found)"}, nil
	}
	return &pb.Response{Success: true, Value: "(1 pair(s) found)"}, nil
}

// Removes a key in a namespace and returns the KVP, if present
func (s *Server) Unset(ctx context.Context, in *pb.Key) (*pb.KeyValuePair, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}
	newKey := token.Username + "." + token.Namespace + "." + in.Key
	value, ok := s.Data.Pop(newKey)
	if !ok {
		return nil, KVPMissingErr
	}
	return &pb.KeyValuePair{Key: in.Key, Value: value.(string)}, nil
}

// Retrieves an element from a namespace under given key
func (s *Server) Get(ctx context.Context, in *pb.Key) (*pb.KeyValuePair, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}
	newKey := token.Username + "." + token.Namespace + "." + in.Key
	value, ok := s.Data.Get(newKey)
	if !ok {
		return nil, KVPMissingErr
	}
	return &pb.KeyValuePair{Key: in.Key, Value: value.(string)}, nil
}

// TODO ========================================================================

// Returns the total number of key-value pairs in a namespace
func (s *Server) Count(ctx context.Context, in *google_protobuf.Empty) (*pb.CountResponse, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}

	newKey := token.Username + "." + token.Namespace
	// TODO: get all the keys, retrieve all keys starting with newKey (inefficient)
	// then get count
	fmt.Println(newKey)
	return nil, nil
}

// Retrieve all keys in a namespace
func (s *Server) ShowKeys(ctx context.Context, in *google_protobuf.Empty) (*pb.ShowKeysResponse, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}

	newKey := token.Username + "." + token.Namespace
	// TODO: get all the keys, retrieve all keys starting with newKey (inefficient)
	fmt.Println(newKey)
	return nil, nil
}

// Retrieve all key-value pairs in a namespace
func (s *Server) ShowData(ctx context.Context, in *google_protobuf.Empty) (*pb.ShowDataResponse, error) {
	token, err := verifyToken(ctx)
	if err != nil {
		return nil, err
	}

	newKey := token.Username + "." + token.Namespace
	// TODO: get all the keys, retrieve all key-value pairs starting with newKey (inefficient)
	fmt.Println(newKey)
	return nil, nil
}

// Retrieve all namespaces in the key-value store that belongs to the user
// NOTE: No token needed
func (s *Server) ShowNamespaces(ctx context.Context, in *google_protobuf.Empty) (*pb.ShowNamespacesResponse, error) {
	// TODO
	return nil, nil
}

// Changes the current namespace, returns a token that must be used for subsequent requests
// NOTE: No token needed
func (s *Server) UseNamespace(ctx context.Context, in *pb.Namespace) (*pb.NamespaceResponse, error) {
	// TODO: Verify if user has permission to the namespace
	// verifies that namespace is alphanumeric
	re := regexp.MustCompile(`[a-zA-Z]`)
	match := re.FindStringSubmatch(in.Namespace)
	if len(match) == 0 {
		return nil, InvalidNamespaceErr
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, EmptyMetadataErr // should not occur
	}
	// initialize token
	claims := Token{
		md["username"][0],
		in.Namespace,
		jwt.StandardClaims{
			Issuer: "keev",
		},
	}
	// sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(common.JWTSigningToken)
	if err != nil {
		return nil, TokenSigningErr
	}
	return &pb.NamespaceResponse{Token: ss}, nil
}
