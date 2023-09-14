package http

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	ByteUnitKB = "KB"
	ByteUnitMB = "MB"
)

type SyncResults struct {
	Url      string
	Body     []byte
	Err      error
	Response *http.Response
	SendCost time.Duration
	SyncCost time.Duration
	Extra    map[string]interface{}
}

// ServerConfig server config
type ServerConfig struct {
	Address            string        `yaml:"address" json:"address" default:":80"`
	Concurrency        int           `yaml:"concurrency" json:"concurrency"`
	DisableKeepalive   bool          `yaml:"disableKeepalive" json:"disableKeepalive"`
	TCPKeepalive       bool          `yaml:"tcpKeepalive" json:"tcpKeepalive"`
	MaxRequestBodySize int           `yaml:"maxRequestBodySize" json:"maxRequestBodySize"`
	ReadTimeout        time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout       time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	IdleTimeout        time.Duration `yaml:"idleTimeout" json:"idleTimeout"`
	utils.Certificate  `yaml:",inline" json:",inline"`
}

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
	SpeedLimit            int
	ByteUnit              string
	SyncMaxConcurrency    int
}

// NewClientOptions creates client options with default values
func NewClientOptions() *ClientOptions {
	return &ClientOptions{
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
	ByteUnit              string        `yaml:"byteUnit" json:"byteUnit" default:"KB"`
	SpeedLimit            int           `yaml:"speedLimit" json:"speedLimit" default:"0"`
	SyncMaxConcurrency    int           `yaml:"syncMaxConcurrency" json:"syncMaxConcurrency" default:"0"`
	utils.Certificate     `yaml:",inline" json:",inline"`
}

// ToClientOptions converts client config to client options
func (cc ClientConfig) ToClientOptions() (*ClientOptions, error) {
	tlsConfig, err := utils.NewTLSConfigClient(cc.Certificate)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &ClientOptions{
		Address:               cc.Address,
		Timeout:               cc.Timeout,
		TLSConfig:             tlsConfig,
		KeepAlive:             cc.KeepAlive,
		MaxIdleConns:          cc.MaxIdleConns,
		IdleConnTimeout:       cc.IdleConnTimeout,
		TLSHandshakeTimeout:   cc.TLSHandshakeTimeout,
		ExpectContinueTimeout: cc.ExpectContinueTimeout,
		SpeedLimit:            cc.SpeedLimit,
		ByteUnit:              cc.ByteUnit,
		SyncMaxConcurrency:    cc.SyncMaxConcurrency,
	}, nil
}

// ToClientOptionsWithPassphrase converts client config to client options with passphrase
func (cc ClientConfig) ToClientOptionsWithPassphrase() (*ClientOptions, error) {
	tlsConfig, err := utils.NewTLSConfigClientWithPassphrase(cc.Certificate)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &ClientOptions{
		Address:               cc.Address,
		Timeout:               cc.Timeout,
		TLSConfig:             tlsConfig,
		KeepAlive:             cc.KeepAlive,
		MaxIdleConns:          cc.MaxIdleConns,
		IdleConnTimeout:       cc.IdleConnTimeout,
		TLSHandshakeTimeout:   cc.TLSHandshakeTimeout,
		ExpectContinueTimeout: cc.ExpectContinueTimeout,
		SpeedLimit:            cc.SpeedLimit,
		ByteUnit:              cc.ByteUnit,
		SyncMaxConcurrency:    cc.SyncMaxConcurrency,
	}, nil
}
