package link

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// ServerOptions server option
type ServerOptions struct {
	Address              string
	TLSConfig            *tls.Config
	LinkServer           LinkServer
	MaxMessageSize       utils.Size
	MaxConcurrentStreams uint32
}

// ClientOptions client options
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

// NewServerOptions creates client options with default values
func NewServerOptions() ServerOptions {
	return ServerOptions{
		MaxMessageSize: 4 * 1024 * 1024,
	}
}

// NewClientOptions creates client options with default values
func NewClientOptions() ClientOptions {
	return ClientOptions{
		MaxMessageSize:       4 * 1024 * 1024,
		MaxCacheMessages:     10,
		MaxReconnectInterval: 3 * time.Minute,
	}
}

// ClientConfig client config
type ClientConfig struct {
	Address              string        `yaml:"address" json:"address"`
	Timeout              time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	MaxReconnectInterval time.Duration `yaml:"maxReconnectInterval" json:"maxReconnectInterval" default:"3m"`
	MaxMessageSize       utils.Size    `yaml:"maxMessageSize" json:"maxMessageSize" default:"4m"`
	MaxCacheMessages     int           `yaml:"maxCacheMessages" json:"maxCacheMessages" default:"10"`
	DisableAutoAck       bool          `yaml:"disableAutoAck" json:"disableAutoAck"`
	utils.Certificate    `yaml:",inline" json:",inline"`
}

// ToClientOptions converts client config to client options
func (cc ClientConfig) ToClientOptions(obs Observer) (*ClientOptions, error) {
	ops := &ClientOptions{
		Address:              cc.Address,
		MaxMessageSize:       cc.MaxMessageSize,
		MaxCacheMessages:     cc.MaxCacheMessages,
		MaxReconnectInterval: cc.MaxReconnectInterval,
		DisableAutoAck:       cc.DisableAutoAck,
		Observer:             obs,
	}
	if cc.Certificate.Key != "" || cc.Certificate.Cert != "" {
		tlsconfig, err := utils.NewTLSConfigClient(cc.Certificate)
		if err != nil {
			return nil, err
		}
		ops.TLSConfig = tlsconfig
	}
	return ops, nil
}
