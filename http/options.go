package http

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// ClientOptions client options
type ClientOptions struct {
	Address               string
	TLSConfig             *tls.Config
	Timeout               time.Duration
	KeepAlive             time.Duration
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

// NewClientOptions creates client options with default values
func NewClientOptions() ClientOptions {
	return ClientOptions{
		Timeout:               30 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// ClientConfig client config
type ClientConfig struct {
	Address               string        `yaml:"address" json:"address"`
	Timeout               time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive             time.Duration `yaml:"keepalive" json:"keepalive" default:"30s"`
	MaxIdleConns          int           `yaml:"maxIdleConns" json:"maxIdleConns" default:"100"`
	IdleConnTimeout       time.Duration `yaml:"idleConnTimeout" json:"idleConnTimeout" default:"90s"`
	TLSHandshakeTimeout   time.Duration `yaml:"tlsHandshakeTimeout" json:"tlsHandshakeTimeout" default:"10s"`
	ExpectContinueTimeout time.Duration `yaml:"expectContinueTimeout" json:"expectContinueTimeout" default:"1s"`
	utils.Certificate     `yaml:",inline" json:",inline"`
}

// ToClientOptions converts client config to client options
func (cc ClientConfig) ToClientOptions() (*ClientOptions, error) {
	ops := &ClientOptions{
		Address:               cc.Address,
		Timeout:               cc.Timeout,
		KeepAlive:             cc.KeepAlive,
		MaxIdleConns:          cc.MaxIdleConns,
		IdleConnTimeout:       cc.IdleConnTimeout,
		TLSHandshakeTimeout:   cc.TLSHandshakeTimeout,
		ExpectContinueTimeout: cc.ExpectContinueTimeout,
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
