package context

import (
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
)

// ServiceConfig base config of service
type ServiceConfig struct {
	HTTP   http.ClientConfig `yaml:"http,omitempty" json:"http,omitempty"`
	MQTT   mqtt.ClientConfig `yaml:"mqtt,omitempty" json:"mqtt,omitempty"`
	Logger log.Config        `yaml:"logger,omitempty" json:"logger,omitempty"`
}
