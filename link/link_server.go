package link

import (
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Call message handler
type Call func(context.Context, *Message) (*Message, error)

// Talk stream message handler
type Talk func(Link_TalkServer) error

// LServer Link server to handle message
type LServer struct {
	addr string
	cfg  LServerConfig
	svr  *grpc.Server
	call Call
	talk Talk
}

// NewLServer creates a new Link server
func NewLServer(c LServerConfig, call Call, talk Talk) (*LServer, error) {
	lis, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{}
	if c.Certificate.Cert != "" && c.Certificate.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(c.Certificate.Cert, c.Certificate.Key)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}
	svr := grpc.NewServer(opts...)
	s := &LServer{
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
func (s *LServer) Call(c context.Context, m *Message) (*Message, error) {
	if s.call == nil {
		return nil, fmt.Errorf("handle not implemented")
	}
	if authResult, err := authenticate(c, s.cfg.Account); !authResult {
		return nil, err
	}
	return s.call(c, m)
}

// Talk stream message handler
func (s *LServer) Talk(stream Link_TalkServer) error {
	if authResult, err := authenticate(stream.Context(), s.cfg.Account); !authResult {
		return err
	}
	return s.talk(stream)
}

// Close closes server
func (s *LServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}

func authenticate(c context.Context, a Account) (bool, error) {
	if len(a.Username) > 0 {
		md, ok := metadata.FromIncomingContext(c)
		if !ok {
			return false, status.Errorf(codes.Unauthenticated, "no metadata")
		}
		// TODO: change user+pwd to token
		var username, password string
		if val, ok := md["username"]; ok {
			username = val[0]
		}
		if val, ok := md["password"]; ok {
			password = val[0]
		}
		if strings.Compare(username, a.Username) != 0 ||
			strings.Compare(password, a.Password) != 0 {
			return false, status.Errorf(codes.Unauthenticated, "username or password not match")
		}
	}
	return true, nil
}
