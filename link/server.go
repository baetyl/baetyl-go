package link

import (
	"fmt"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// all metadata keys
const (
	KeyUsername = "username"
	KeyPassword = "password"
)

// ErrUnauthenticated ErrUnauthenticated
var ErrUnauthenticated = status.Errorf(codes.Unauthenticated, "Username is unauthenticated")

// Authenticator : Authenticate interface
type Authenticator interface {
	Authenticate(context.Context) error
}

// NewServer creates a new grpc server
func NewServer(cfg ServerConfig, auth Authenticator) (*grpc.Server, error) {
	logger := log.With(log.Any("link", "server"))

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(cfg.MaxConcurrent),
		grpc.MaxRecvMsgSize(int(cfg.MaxMessageSize)),
		grpc.MaxSendMsgSize(int(cfg.MaxMessageSize)),
	}
	if cfg.Certificate.Key != "" || cfg.Certificate.Cert != "" {
		tlsCfg, err := utils.NewTLSConfigServer(cfg.Certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.Creds(creds))
	}
	if auth != nil {
		ui := func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			if ent := logger.Check(log.DebugLevel, "server received a message"); ent != nil {
				ent.Write(log.Any("message", fmt.Sprintf("%v", req)))
			}
			err := auth.Authenticate(ctx)
			if err != nil {
				logger.Error("Unauthenticated")
				return nil, err
			}
			return handler(ctx, req)
		}
		si := func(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			logger.Debug("server accepted a stream")
			err := auth.Authenticate(ss.Context())
			if err != nil {
				logger.Error("Unauthenticated")
				return err
			}
			return handler(srv, ss)
		}
		opts = append(opts, grpc.UnaryInterceptor(ui), grpc.StreamInterceptor(si))
	}

	svr := grpc.NewServer(opts...)
	reflection.Register(svr)
	return svr, nil
}
