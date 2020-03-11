package mqtt

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// ClientOptions mqtt client options
type ClientOptions struct {
	Address              string
	Username             string
	Password             string
	TLSConfig            *tls.Config
	ClientID             string
	CleanSession         bool
	Timeout              time.Duration
	KeepAlive            time.Duration
	MaxReconnectInterval time.Duration
	MaxMessageSize       utils.Size
	MaxCacheMessages     int
	DisableAutoAck       bool
	Observer             Observer
}

// NewClientOptions creates mqtt client options with default values
func NewClientOptions() ClientOptions {
	return ClientOptions{
		Timeout:              30 * time.Second,
		KeepAlive:            3 * time.Minute,
		MaxReconnectInterval: 3 * time.Minute,
		MaxMessageSize:       4 * 1024 * 1024,
		MaxCacheMessages:     10,
	}
}
