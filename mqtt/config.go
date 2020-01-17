package mqtt

import (
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// QOSTopic topic and qos
type QOSTopic struct {
	QOS   uint32 `yaml:"qos" json:"qos" validate:"min=0, max=1"`
	Topic string `yaml:"topic" json:"topic" validate:"nonzero"`
}

// ServerConfig mqtt server config
type ServerConfig struct {
	Addresses   []string          `yaml:"addresses" json:"addresses"`
	Certificate utils.Certificate `yaml:",inline" json:",inline"`
}

// ClientConfig mqtt client config
type ClientConfig struct {
	Address        string            `yaml:"address" json:"address"`
	Username       string            `yaml:"username" json:"username"`
	Password       string            `yaml:"password" json:"password"`
	Certificate    utils.Certificate `yaml:",inline" json:",inline"`
	ClientID       string            `yaml:"clientid" json:"clientid"`
	CleanSession   bool              `yaml:"cleansession" json:"cleansession"`
	KeepAlive      time.Duration     `yaml:"keepalive" json:"keepalive"` // keepalive not enabled by default
	Timeout        time.Duration     `yaml:"timeout" json:"timeout" default:"30s"`
	Interval       time.Duration     `yaml:"interval" json:"interval" default:"2m"`
	BufferSize     int               `yaml:"buffersize" json:"buffersize" default:"10"`
	DisableAutoAck bool              `yaml:"disableAutoAck" json:"disableAutoAck"`
}
