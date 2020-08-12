package context

import (
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/utils"
)

// SystemConfig config of baetyl system
type SystemConfig struct {
	Certificate utils.Certificate `yaml:"cert,omitempty" json:"cert,omitempty" default:"{\"ca\":\"var/lib/baetyl/system/certs/ca.pem\",\"key\":\"var/lib/baetyl/system/certs/key.pem\",\"cert\":\"var/lib/baetyl/system/certs/crt.pem\"}"`
	Function    http.ClientConfig `yaml:"function,omitempty" json:"function,omitempty"`
	Broker      mqtt.ClientConfig `yaml:"broker,omitempty" json:"broker,omitempty"`
	Logger      log.Config        `yaml:"logger,omitempty" json:"logger,omitempty"`
}
