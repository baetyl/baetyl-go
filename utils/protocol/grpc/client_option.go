package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ClientOption a builder for build grpc server option
type ClientOption struct {
	opts []grpc.DialOption
}

// CustomCred for Custom Credential, implement GetRequestMetadata & RequireTransportSecurity
type CustomCred struct {
	Username string
	Password string
}

// GetRequestMetadata & RequireTransportSecurity for Custom Credential
// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (c *CustomCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (c *CustomCred) RequireTransportSecurity() bool {
	return len(c.Username) > 0
}

// Create create []ClientOption
func (c *ClientOption) Create() *ClientOption {
	c.opts = []grpc.DialOption{}
	return c
}

// Build return DialOption
func (c *ClientOption) Build() []grpc.DialOption {
	return c.opts
}

// Option set other option
func (c *ClientOption) Option(opt grpc.DialOption) *ClientOption {
	c.opts = append(c.opts, opt)
	return c
}

// CredsFromFile set TLS by file
// Name : serverNameOverride, same to CommonName in server.pem
func (c *ClientOption) CredsFromFile(cert, name string) *ClientOption {
	if cert != "" {
		creds, err := credentials.NewClientTLSFromFile(cert, name)
		if err != nil {
			fmt.Printf("ClientOption CredsFromFile NewClientTLSFromFile err = %s\n", err.Error())
			return c
		}
		c.opts = append(c.opts, grpc.WithTransportCredentials(creds))
	}
	return c
}

// Creds set TLS
func (c *ClientOption) Creds(creds credentials.TransportCredentials) *ClientOption {
	c.opts = append(c.opts, grpc.WithTransportCredentials(creds))
	return c
}

// CustomCred for Custom Credential
func (c *ClientOption) CustomCred(cc *CustomCred) *ClientOption {
	c.opts = append(c.opts, grpc.WithPerRPCCredentials(cc))
	return c
}

// Insecure disables transport security for this ClientConn
func (c *ClientOption) Insecure(flag bool) *ClientOption {
	if flag {
		c.opts = append(c.opts, grpc.WithInsecure())
	}
	return c
}
