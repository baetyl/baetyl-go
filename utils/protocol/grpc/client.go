package grpc

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewClientConnect create a new client connect of grpc server
func NewClientConnect(address string, timeout time.Duration, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cel := context.WithTimeout(context.Background(), timeout)
	defer cel()

	return grpc.DialContext(ctx, address, opts...)
}
