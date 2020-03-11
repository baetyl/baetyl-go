package context

import (
	"time"

	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
	"github.com/baetyl/baetyl-go/utils"
)

// ServiceConfig base config of service
type ServiceConfig struct {
	Mqtt   MQTTClientConfig `yaml:"mqtt" json:"mqtt"`
	Link   LinkClientConfig `yaml:"link" json:"link"`
	Logger log.Config       `yaml:"logger" json:"logger"`
}

// MQTTClientConfig mqtt client config
type MQTTClientConfig struct {
	Address              string            `yaml:"address" json:"address"`
	Username             string            `yaml:"username" json:"username"`
	Password             string            `yaml:"password" json:"password"`
	Certificate          utils.Certificate `yaml:",inline" json:",inline"`
	ClientID             string            `yaml:"clientid" json:"clientid"`
	CleanSession         bool              `yaml:"cleansession" json:"cleansession"`
	Timeout              time.Duration     `yaml:"timeout" json:"timeout" default:"30s"`
	KeepAlive            time.Duration     `yaml:"keepalive" json:"keepalive" default:"3m"`
	MaxReconnectInterval time.Duration     `yaml:"maxReconnectInterval" json:"maxReconnectInterval" default:"3m"`
	MaxCacheMessages     int               `yaml:"maxCacheMessages" json:"maxCacheMessages" default:"10"`
	DisableAutoAck       bool              `yaml:"disableAutoAck" json:"disableAutoAck"`
	Subscriptions        Subscriptions     `yaml:"subscriptions" json:"subscriptions" default:"[]"`
}

// LinkClientConfig link client config
type LinkClientConfig struct {
	Address              string            `yaml:"address" json:"address"`
	Certificate          utils.Certificate `yaml:",inline" json:",inline"`
	Timeout              time.Duration     `yaml:"timeout" json:"timeout" default:"30s"`
	MaxReconnectInterval time.Duration     `yaml:"maxReconnectInterval" json:"maxReconnectInterval" default:"3m"`
	MaxMessageSize       utils.Size        `yaml:"maxMessageSize" json:"maxMessageSize" default:"4m"`
	MaxCacheMessages     int               `yaml:"maxCacheMessages" json:"maxCacheMessages" default:"10"`
	DisableAutoAck       bool              `yaml:"disableAutoAck" json:"disableAutoAck"`
}

// Subscriptions subscriptions
type Subscriptions []QOSTopic

// QOSTopic topic and qos
type QOSTopic struct {
	QOS   uint32 `yaml:"qos" json:"qos" validate:"min=0, max=1"`
	Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
}

// ToClientOptions converts client config to client options
func (cc MQTTClientConfig) ToClientOptions(obs mqtt.Observer) (*mqtt.ClientOptions, error) {
	ops := &mqtt.ClientOptions{
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

// ToClientOptions converts client config to client options
func (cc LinkClientConfig) ToClientOptions(obs link.Observer) (*link.ClientOptions, error) {
	ops := &link.ClientOptions{
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

// ToMQTTSubscriptions converts to mqtt subscriptions
func (ss Subscriptions) ToMQTTSubscriptions() []mqtt.Subscription {
	var subs []mqtt.Subscription
	for _, topic := range ss {
		subs = append(subs, mqtt.Subscription{Topic: topic.Topic, QOS: mqtt.QOS(topic.QOS)})
	}
	return subs
}
