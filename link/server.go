package link

import (
	"github.com/baetyl/baetyl-go/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func NewServer(c ServerConfig) (*grpc.Server, error) {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(c.Concurrent.Max),
		grpc.MaxRecvMsgSize(int(c.MaxSize)),
		grpc.MaxSendMsgSize(int(c.MaxSize)),
	}
	tlsCfg, err := utils.NewTLSConfigServer(&c.Certificate)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.Creds(creds))
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if len(c.Username) > 0 && len(c.Password) > 0 {
			auth := &AuthPassword{
				Username: c.Username,
				Password: c.Password,
			}
			err = auth.Authenticate(ctx)
			if err != nil {
				return resp, err
			}
		}
		// todo auth token
		return handler(ctx, req)
	}
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	svr := grpc.NewServer(opts...)
	reflection.Register(svr)
	return svr, nil
}
