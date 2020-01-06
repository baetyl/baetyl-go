package api

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"google.golang.org/grpc/metadata"
)

// Master interface
type Master interface {
	Auth(u, p string) bool
}

// NewServer creates a new server
func NewServer(conf ServerConfig, m Master) (*Server, error) {
	utils.SetDefaults(&conf)
	svr, err := link.NewServer(link.ServerConfig(conf), Authenticator{m: m})
	if err != nil {
		return nil, err
	}
	return &Server{conf: conf, svr: svr}, nil
}

// RegisterKVService register kv service
func (s *Server) RegisterKVService(server KVServiceServer) {
	RegisterKVServiceServer(s.svr, server)
}

// Start start api server
func (s *Server) Start() error {
	logger := log.With(log.Any("api", "server"))

	uri, err := utils.ParseURL(s.conf.Address)
	if err != nil {
		return err
	}

	if uri.Scheme == "unix" {
		if err := syscall.Unlink(uri.Host); err != nil {
			logger.Error("failed to unlink sock file", log.Error(err))
		}
		dir := filepath.Dir(uri.Host)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			logger.Error("failed to make directory", log.Any("directory", dir), log.Error(err))
		}
	}
	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return err
	}
	logger.Info("api server is listening", log.Any("address", s.conf.Address))
	go func() {
		if err := s.svr.Serve(listener); err != nil {
			logger.Info("api server shutdown", log.Error(err))
		}
	}()
	return nil
}

// Close closes api server
func (s *Server) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}

type Authenticator struct {
	m Master
}

func (a Authenticator) Authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return link.ErrUnauthenticated
	}
	var u, p string
	if val, ok := md[link.KeyUsername]; ok {
		u = val[0]
	}
	if val, ok := md[link.KeyPassword]; ok {
		p = val[0]
	}
	if ok := a.m.Auth(u, p); !ok {
		return link.ErrUnauthenticated
	}
	return nil
}
