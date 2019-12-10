package mqtt

import (
	"time"

	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl-go/utils"
)

// The Dialer handles connecting to a server and creating a connection.
type Dialer struct {
	*transport.Dialer
}

// NewDialer returns a new Dialer.
func NewDialer(c utils.Certificate, t time.Duration) (*Dialer, error) {
	tls, err := utils.NewTLSConfigClient(c)
	if err != nil {
		return nil, err
	}
	return &Dialer{Dialer: transport.NewDialer(transport.DialConfig{TLSConfig: tls, Timeout: t})}, nil
}
