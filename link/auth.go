package link

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

// Authenticator : Authenticate interface
type Authenticator interface {
	Authenticate(context.Context) error
}

// AuthPassword : authenticate by token
type AuthToken struct {
	Token string
}

func (a *AuthToken) Authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "no metadata")
	}
	// todo auth by token
	fmt.Printf("todo AuthToken %v", md)
	return nil
}

// AuthPassword : authenticate by username and password
type AuthPassword struct {
	Username string
	Password string
}

func (a *AuthPassword) Authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "no metadata")
	}
	var u, p string
	if val, ok := md[KeyUsername]; ok {
		u = val[0]
	}
	if val, ok := md[KeyPassword]; ok {
		p = val[0]
	}
	if strings.Compare(u, a.Username) != 0 ||
		strings.Compare(p, a.Password) != 0 {
		return status.Errorf(codes.Unauthenticated, "username or password not match")
	}
	return nil
}

// CustomCred implement GetRequestMetadata & RequireTransportSecurity
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
	return false
}
