package main

import "errors"

var (
	KVPExistsErr        = errors.New("key already exists")
	KVPMissingErr       = errors.New("key does not exist")
	MissingTokenErr     = errors.New("missing token for namespace, use Use() to set a namespace")
	InvalidTokenErr     = errors.New("invalid token for namespace, use Use() to set a namespace")
	EmptyMetadataErr    = errors.New("missing metadata, please login again")
	TokenSigningErr     = errors.New("unable to sign token")
	InvalidNamespaceErr = errors.New("invalid namespace, alphanumerics only")
	AccessDeniedErr     = errors.New("access denied: invalid username or password")
)
