package link

import (
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// Account authentication information
type Account struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

type Auth struct {
	Account           `yaml:"address" json:"address" default: "0.0.0.0"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// ServerConfig link server config
type ServerConfig struct {
	MaxMessageSize utils.Size `yaml:"maxMessageSize" json:"maxMessageSize"`
	Concurrent     struct {
		Max uint32 `yaml:"max" json:"max" default:"{\"max\":4194304}`
	} `yaml:"concurrent" json:"concurrent"`
	Auth `yaml:"auth" json:"auth"`
}

// ClientConfig link client config
type ClientConfig struct {
	Address        string        `yaml:"address" json:"address" default: "0.0.0.0"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	MaxMessageSize utils.Size    `yaml:"maxMessageSize" json:"maxMessageSize"`
	DisableAutoAck bool          `yaml:"disableAutoAck" json:"disableAutoAck"`
	Auth           `yaml:"auth" json:"auth"`
}
