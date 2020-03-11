package link

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// ServerOptions link server option
type ServerOptions struct {
	Address              string
	TLSConfig            *tls.Config
	LinkServer           LinkServer
	MaxMessageSize       utils.Size
	MaxConcurrentStreams uint32
}

// ClientOptions link client options
type ClientOptions struct {
	Address              string
	TLSConfig            *tls.Config
	MaxMessageSize       utils.Size
	MaxCacheMessages     int
	MaxReconnectInterval time.Duration
	DisableAutoAck       bool
	Observer             Observer
}

// Observer message observer interface
type Observer interface {
	OnMsg(*Message) error
	OnAck(*Message) error
	OnErr(error)
}

// NewServerOptions creates link client options with default values
func NewServerOptions() ServerOptions {
	return ServerOptions{
		MaxMessageSize: 4 * 1024 * 1024,
	}
}

// NewClientOptions creates link client options with default values
func NewClientOptions() ClientOptions {
	return ClientOptions{
		MaxMessageSize:       4 * 1024 * 1024,
		MaxCacheMessages:     10,
		MaxReconnectInterval: 3 * time.Minute,
	}
}
