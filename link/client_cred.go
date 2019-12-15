package link

import (
	"context"
)

// MD custom metadata
type MD map[string]string

// GetRequestMetadata & RequireTransportSecurity for Custom Credential
// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (md MD) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return md, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (md MD) RequireTransportSecurity() bool {
	return false
}
