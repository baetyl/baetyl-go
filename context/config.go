package context

import (
	"github.com/baetyl/baetyl-go.v2/http"
	"github.com/baetyl/baetyl-go.v2/log"
	"github.com/baetyl/baetyl-go.v2/mqtt"
)

// ServiceConfig base config of service
type ServiceConfig struct {
	HTTP   http.ClientConfig `yaml:"http,omitempty" json:"http,omitempty"`
	MQTT   mqtt.ClientConfig `yaml:"mqtt,omitempty" json:"mqtt,omitempty"`
	Logger log.Config        `yaml:"logger,omitempty" json:"logger,omitempty"`
}
