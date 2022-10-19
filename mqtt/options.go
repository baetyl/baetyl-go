package mqtt

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

// ClientOptions client options
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
	Subscriptions        []Subscription
	DisableAutoAck       bool
}

// NewClientOptions creates client options with default values
func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		Timeout:              30 * time.Second,
		KeepAlive:            3 * time.Minute,
		MaxReconnectInterval: 3 * time.Minute,
		MaxMessageSize:       4 * 1024 * 1024,
		MaxCacheMessages:     10,
	}
}

// QOSTopic topic and qos
type QOSTopic struct {
	QOS   uint32 `yaml:"qos" json:"qos" binding:"min=0,max=1"`
	Topic string `yaml:"topic" json:"topic" binding:"nonzero"`
}

// ClientConfig client config
type ClientConfig struct {
	Address              string        `yaml:"address" json:"address"`
	Username             string        `yaml:"username" json:"username"`
	Password             string        `yaml:"password" json:"password"`
	ClientID             string        `yaml:"clientid" json:"clientid"`
	CleanSession         bool          `yaml:"cleansession" json:"cleansession"`
	Timeout              time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive            time.Duration `yaml:"keepalive" json:"keepalive" default:"30s"`
	MaxReconnectInterval time.Duration `yaml:"maxReconnectInterval" json:"maxReconnectInterval" default:"3m"`
	MaxCacheMessages     int           `yaml:"maxCacheMessages" json:"maxCacheMessages" default:"10"`
	DisableAutoAck       bool          `yaml:"disableAutoAck" json:"disableAutoAck"`
	Subscriptions        []QOSTopic    `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	utils.Certificate    `yaml:",inline" json:",inline"`
}

// ToClientOptions converts client config to client options
func (cc ClientConfig) ToClientOptions() (*ClientOptions, error) {
	ops := &ClientOptions{
		Address:              cc.Address,
		Username:             cc.Username,
		Password:             cc.Password,
		ClientID:             cc.ClientID,
		CleanSession:         cc.CleanSession,
		Timeout:              cc.Timeout,
		KeepAlive:            cc.KeepAlive,
		MaxReconnectInterval: cc.MaxReconnectInterval,
		MaxCacheMessages:     cc.MaxCacheMessages,
		DisableAutoAck:       cc.DisableAutoAck,
	}
	if cc.Certificate.Key != "" || cc.Certificate.Cert != "" {
		tlsconfig, err := utils.NewTLSConfigClient(cc.Certificate)
		if err != nil {
			return nil, errors.Trace(err)
		}
		ops.TLSConfig = tlsconfig
	}
	for _, topic := range cc.Subscriptions {
		ops.Subscriptions = append(ops.Subscriptions, Subscription{Topic: topic.Topic, QOS: QOS(topic.QOS)})
	}
	return ops, nil
}
