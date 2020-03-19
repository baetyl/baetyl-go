package mqtt

import (
	"crypto/tls"
	"time"

	"github.com/baetyl/baetyl-go/utils"
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
	DisableAutoAck       bool
	Observer             Observer
}

// NewClientOptions creates client options with default values
func NewClientOptions() ClientOptions {
	return ClientOptions{
		Timeout:              30 * time.Second,
		KeepAlive:            3 * time.Minute,
		MaxReconnectInterval: 3 * time.Minute,
		MaxMessageSize:       4 * 1024 * 1024,
		MaxCacheMessages:     10,
	}
}

// QOSTopic topic and qos
type QOSTopic struct {
	QOS   uint32 `yaml:"qos" json:"qos" validate:"min=0, max=1"`
	Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
}

// Subscriptions subscriptions
type Subscriptions []QOSTopic

// ToMQTTSubscriptions converts to mqtt subscriptions
func (ss Subscriptions) ToMQTTSubscriptions() []Subscription {
	var subs []Subscription
	for _, topic := range ss {
		subs = append(subs, Subscription{Topic: topic.Topic, QOS: QOS(topic.QOS)})
	}
	return subs
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
	Subscriptions        Subscriptions `yaml:"subscriptions" json:"subscriptions" default:"[]"`
	utils.Certificate    `yaml:",inline" json:",inline"`
}

// ToClientOptions converts client config to client options
func (cc ClientConfig) ToClientOptions(obs Observer) (*ClientOptions, error) {
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
