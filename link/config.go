package link

import (
	"time"

	"github.com/baetyl/baetyl-go/utils"
)

// ServerConfig link server config
type ServerConfig struct {
	Address           string     `yaml:"address" json:"address"`
	MaxConcurrent     uint32     `yaml:"MaxConcurrent" json:"maxConcurrent"`
	MaxMessageSize    utils.Size `yaml:"maxMessageSize" json:"maxMessageSize" default:"4m"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// ClientConfig link client config
type ClientConfig struct {
	Address           string        `yaml:"address" json:"address"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Interval          time.Duration `yaml:"interval" json:"interval" default:"2m"`
	MaxMessageSize    utils.Size    `yaml:"maxMessageSize" json:"maxMessageSize" default:"4m"`
	MaxCacheMessages  int           `yaml:"maxCacheMessages" json:"maxCacheMessages" default:"10"`
	Username          string        `yaml:"username" json:"username"`
	Password          string        `yaml:"password" json:"password"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
