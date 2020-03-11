package link

import (
	"crypto/tls"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server is a gRPC server to serve RPC requests.
type Server = grpc.Server

// Launch launches a link server
func Launch(op ServerOptions) (*Server, error) {
	var err error
	var l net.Listener
	if op.TLSConfig == nil {
		l, err = net.Listen("tcp", op.Address)
	} else {
		l, err = tls.Listen("tcp", op.Address, op.TLSConfig)
	}
	if err != nil {
		return nil, err
	}
	gops := []grpc.ServerOption{}
	if op.TLSConfig != nil {
		gops = append(gops, grpc.Creds(credentials.NewTLS(op.TLSConfig)))
	}
	if op.MaxMessageSize > 0 {
		gops = append(gops, grpc.MaxRecvMsgSize(int(op.MaxMessageSize)))
		gops = append(gops, grpc.MaxSendMsgSize(int(op.MaxMessageSize)))
	}
	if op.MaxConcurrentStreams > 0 {
		gops = append(gops, grpc.MaxConcurrentStreams(op.MaxConcurrentStreams))
	}
	svr := grpc.NewServer(gops...)
	RegisterLinkServer(svr, op.LinkServer)
	go svr.Serve(l)
	return svr, nil
}
