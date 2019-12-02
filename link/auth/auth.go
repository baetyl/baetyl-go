package auth

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	KeyUsername = "username"
	KeyPassword = "password"
	KeyToken    = "token"
	KeyAK       = "ak"
	KeySK       = "sk"
)

type Auth interface {
	AuthToken(c context.Context, token string) (bool, error)
	AuthPassword(c context.Context, username, password string) (bool, error)
}

// AuthPassword grpc auth by username and password
func AuthPassword(c context.Context, username, password string) (bool, error) {
	md, ok := metadata.FromIncomingContext(c)
	if !ok {
		return false, status.Errorf(codes.Unauthenticated, "no metadata")
	}
	var u, p string
	if val, ok := md[KeyUsername]; ok {
		u = val[0]
	}
	if val, ok := md[KeyPassword]; ok {
		p = val[0]
	}
	if strings.Compare(u, username) != 0 ||
		strings.Compare(p, password) != 0 {
		return false, status.Errorf(codes.Unauthenticated, "username or password not match")
	}
	return true, nil
}

// AuthToken auth by token
func AuthToken(c context.Context, token string) (bool, error) {
	md, ok := metadata.FromIncomingContext(c)
	if !ok {
		return false, status.Errorf(codes.Unauthenticated, "no metadata")
	}
	// todo auth by token
	fmt.Printf("todo AuthToken %v", md)
	return true, nil
}

type CustomCred struct {
	Data map[string]string
}

// GetRequestMetadata & RequireTransportSecurity for Custom Credential
// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (c *CustomCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return c.Data, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (c *CustomCred) RequireTransportSecurity() bool {
	return len(c.Data) != 0
}
