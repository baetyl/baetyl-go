package link

import (
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

type Length struct {
	Max int64 `yaml:"max" json:"max"`
}

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
	Address string `yaml:"address" json:"address" default: "0.0.0.0"`
	Message struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	Concurrent struct {
		Max uint32 `yaml:"max" json:"max" default:"{\"max\":4194304}`
	} `yaml:"concurrent" json:"concurrent"`
	Auth `yaml:"auth" json:"auth"`
}

// ClientConfig link client config
type ClientConfig struct {
	Address string        `yaml:"address" json:"address" default: "0.0.0.0"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Message struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	Auth `yaml:"auth" json:"auth"`
}
