package baetyl

import (
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Callback message handler
type Callback func(context.Context, *Message, ...grpc.CallOption) (*Message, error)

// Talk stream message handler
type Talk func(Contact_TalkServer) error

// CServer Contact server to handle message
type CServer struct {
	addr string
	cfg  ContactServerConfig
	svr  *grpc.Server
	call Callback
	talk Talk
}

// NewCServer creates a new Contact server
func NewCServer(c ContactServerConfig, call Callback, talk Talk) (*CServer, error) {
	lis, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, err
	}
	tls, err := utils.NewTLSServerConfig(c.Auth.Certificate)
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(c.Concurrent.Max),
		grpc.MaxRecvMsgSize(int(c.Message.Length.Max)),
		grpc.MaxSendMsgSize(int(c.Message.Length.Max)),
	}
	if tls != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tls)))
	}
	svr := grpc.NewServer(opts...)
	s := &CServer{
		addr: lis.Addr().String(),
		cfg:  c,
		svr:  svr,
		call: call,
		talk: talk,
	}
	RegisterContactServer(svr, s)
	reflection.Register(svr)
	go s.svr.Serve(lis)
	return s, nil
}

// Callback handles message
func (s *CServer) Callback(c context.Context, m *Message) (*Message, error) {
	if s.call == nil {
		return nil, fmt.Errorf("handle not implemented")
	}
	if authResult, err := authenticate(c, s.cfg.Auth); !authResult {
		return nil, err
	}
	return s.call(c, m)
}

// Talk stream message handler
func (s *CServer) Talk(stream Contact_TalkServer) error {
	if authResult, err := authenticate(stream.Context(), s.cfg.Auth); !authResult {
		return err
	}
	return s.talk(stream)
}

// Close closes server
func (s *CServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}

func authenticate(c context.Context, a Auth) (bool, error) {
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
