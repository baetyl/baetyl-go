package api

import (
	"github.com/baetyl/baetyl-go/link"
	"google.golang.org/grpc"
)

type ServerConfig link.ServerConfig
type ClientConfig link.ClientConfig

// Server server to handle grpc message
type Server struct {
	conf ServerConfig
	svr  *grpc.Server
}

// Client server to handle grpc message
type Client struct {
	conf ClientConfig
	conn *grpc.ClientConn
	KV   KVServiceClient
}
