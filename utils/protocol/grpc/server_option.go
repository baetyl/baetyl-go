package grpc

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServerOption a builder for build grpc server option
type ServerOption struct {
	opts []grpc.ServerOption
}

// Create create []ServerOption
func (s *ServerOption) Create() *ServerOption {
	s.opts = []grpc.ServerOption{}
	return s
}

// Build return ServerOption
func (s *ServerOption) Build() []grpc.ServerOption {
	return s.opts
}

// Option set other option
func (s *ServerOption) Option(opt grpc.ServerOption) *ServerOption {
	s.opts = append(s.opts, opt)
	return s
}

// CredsFromFile set TLS from file
func (s *ServerOption) CredsFromFile(cert, key string) *ServerOption {
	if cert != "" && key != "" {
		creds, err := credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			fmt.Printf("ServerOption CredsFromFile NewServerTLSFromFile err = %s\n", err.Error())
			return s
		}
		s.opts = append(s.opts, grpc.Creds(creds))
	}
	return s
}

// Creds set TLS
func (s *ServerOption) Creds(creds credentials.TransportCredentials) *ServerOption {
	s.opts = append(s.opts, grpc.Creds(creds))
	return s
}

// MaxConcurrentStreams set max concurrent stream
func (s *ServerOption) MaxConcurrentStreams(max uint32) *ServerOption {
	s.opts = append(s.opts, grpc.MaxConcurrentStreams(max))
	return s
}

// MaxRecvMsgSize set max receive message size
func (s *ServerOption) MaxRecvMsgSize(size int) *ServerOption {
	s.opts = append(s.opts, grpc.MaxRecvMsgSize(size))
	return s
}

// MaxSendMsgSize set max send message size
func (s *ServerOption) MaxSendMsgSize(size int) *ServerOption {
	s.opts = append(s.opts, grpc.MaxSendMsgSize(size))
	return s
}

// ConnectionTimeout set timeout
func (s *ServerOption) ConnectionTimeout(d time.Duration) *ServerOption {
	s.opts = append(s.opts, grpc.ConnectionTimeout(d))
	return s
}
