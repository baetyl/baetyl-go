package link

import (
	"crypto/tls"
	"net"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server is a gRPC server to serve RPC requests.
type Server = grpc.Server

// Launch launches a link server
func Launch(ops ServerOptions) (*Server, error) {
	// remove tcp/link scheme from address if exists
	addr, err := url.Parse(ops.Address)
	if err == nil && (addr.Scheme == "link" || addr.Scheme == "tcp") {
		ops.Address = addr.Host
	}
	var l net.Listener
	if ops.TLSConfig == nil {
		l, err = net.Listen("tcp", ops.Address)
	} else {
		l, err = tls.Listen("tcp", ops.Address, ops.TLSConfig)
	}
	if err != nil {
		return nil, err
	}
	var gops []grpc.ServerOption
	if ops.TLSConfig != nil {
		gops = append(gops, grpc.Creds(credentials.NewTLS(ops.TLSConfig)))
	}
	if ops.MaxMessageSize > 0 {
		gops = append(gops, grpc.MaxRecvMsgSize(int(ops.MaxMessageSize)))
		gops = append(gops, grpc.MaxSendMsgSize(int(ops.MaxMessageSize)))
	}
	if ops.MaxConcurrentStreams > 0 {
		gops = append(gops, grpc.MaxConcurrentStreams(ops.MaxConcurrentStreams))
	}
	svr := grpc.NewServer(gops...)
	RegisterLinkServer(svr, ops.LinkServer)
	go svr.Serve(l)
	return svr, nil
}
