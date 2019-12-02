package link

import (
	"fmt"
	"net"

	"github.com/baetyl/baetyl-go/link/auth"
	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Call message handler
type Call func(context.Context, *Message) (*Message, error)

// Talk stream message handler
type Talk func(Link_TalkServer) error

// Server Link server to handle message
type Server struct {
	addr string
	cfg  ServerConfig
	svr  *grpc.Server
	call Call
	talk Talk
}

func NewServer(c ServerConfig, call Call, talk Talk) (*Server, error) {
	lis, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(c.Concurrent.Max),
		grpc.MaxRecvMsgSize(int(c.Message.Length.Max)),
		grpc.MaxSendMsgSize(int(c.Message.Length.Max)),
	}
	tlsCfg, err := utils.NewTLSConfigServer(&c.Certificate)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.Creds(creds))
	}
	svr := grpc.NewServer(opts...)
	s := &Server{
		addr: lis.Addr().String(),
		cfg:  c,
		svr:  svr,
		call: call,
		talk: talk,
	}
	RegisterLinkServer(svr, s)
	reflection.Register(svr)
	go s.svr.Serve(lis)
	return s, nil
}

// Call handles message
func (s *Server) Call(c context.Context, m *Message) (*Message, error) {
	if s.call == nil {
		return nil, fmt.Errorf("call handle not implemented")
	}
	if authResult, err := auth.AuthPassword(c, s.cfg.Username, s.cfg.Password); !authResult {
		return nil, err
	}
	return s.call(c, m)
}

// Talk stream message handler
func (s *Server) Talk(stream Link_TalkServer) error {
	if s.talk == nil {
		return fmt.Errorf("talk handle not implemented")
	}
	if authResult, err := auth.AuthPassword(stream.Context(), s.cfg.Username, s.cfg.Password); !authResult {
		return err
	}
	return s.talk(stream)
}

// Close closes server
func (s *Server) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
