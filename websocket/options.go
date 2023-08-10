package websocket

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	ByteUnitKB = "KB"
	ByteUnitMB = "MB"
)

type SyncResults struct {
	Err      error
	SendCost time.Duration
	SyncCost time.Duration
	Extra    map[string]interface{}
}

type ReadMsg struct {
	MsgType int
	Data    []byte
	Err     error
}

type ClientOptions struct {
	Address             string
	Schema              string
	Path                string
	TLSConfig           *tls.Config
	TLSHandshakeTimeout time.Duration
	SyncMaxConcurrency  int
}

type ClientConfig struct {
	Address             string        `yaml:"address" json:"address"`
	Path                string        `yaml:"path" json:"path"`
	Schema              string        `yaml:"schema" json:"schema" default:"ws"`
	IdleConnTimeout     time.Duration `yaml:"idleConnTimeout" json:"idleConnTimeout" default:"90s"`
	TLSHandshakeTimeout time.Duration `yaml:"tlsHandshakeTimeout" json:"tlsHandshakeTimeout" default:"10s"`
	SyncMaxConcurrency  int           `yaml:"syncMaxConcurrency" json:"syncMaxConcurrency" default:"0"`
	utils.Certificate   `yaml:",inline" json:",inline"`
}

func (cc ClientConfig) ToClientOptions() (*ClientOptions, error) {
	tlsConfig, err := utils.NewTLSConfigClient(cc.Certificate)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &ClientOptions{
		Address:             cc.Address,
		Path:                cc.Path,
		Schema:              cc.Schema,
		TLSConfig:           tlsConfig,
		TLSHandshakeTimeout: cc.TLSHandshakeTimeout,
		SyncMaxConcurrency:  cc.SyncMaxConcurrency,
	}, nil
}
