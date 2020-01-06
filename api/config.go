package api

import (
	"github.com/baetyl/baetyl-go/link"
	"google.golang.org/grpc"
)

type ClientConfig link.ClientConfig

// Client server to handle grpc message
type Client struct {
	conf ClientConfig
	conn *grpc.ClientConn
	KV   KVServiceClient
}
