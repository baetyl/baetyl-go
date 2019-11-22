package grpc

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	// NetTCP network tcp
	NetTCP = "tcp"
	// NetTCP4 network tcp4
	NetTCP4 = "tcp4"
	// NetTCP6 network tcp6
	NetTCP6 = "tcp6"
	// NetUnix network unix
	NetUnix = "unix"
	// NetUnixPacket network unixpacket
	NetUnixPacket = "unixpacket"
)

// Register register a grpc service
type Register func(svr *grpc.Server)

// NewServer create grpc server and listen
func NewServer(network, address string,
	opts []grpc.ServerOption,
	register Register) (*grpc.Server, error) {
	lis, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	svr := grpc.NewServer(opts...)
	register(svr)
	reflection.Register(svr)
	go svr.Serve(lis)
	return svr, nil
}

// Authenticate use in rpc function for verify username and password
func Authenticate(c context.Context, username, password string) (bool, error) {
	if len(username) > 0 {
		md, ok := metadata.FromIncomingContext(c)
		if !ok {
			return false, status.Errorf(codes.Unauthenticated, "no metadata")
		}
		var u, p string
		if val, ok := md["username"]; ok {
			u = val[0]
		}
		if val, ok := md["password"]; ok {
			p = val[0]
		}
		if strings.Compare(u, username) != 0 ||
			strings.Compare(p, password) != 0 {
			return false, status.Errorf(codes.Unauthenticated, "username or password not match")
		}
	}
	return true, nil
}
