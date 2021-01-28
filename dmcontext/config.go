package dmcontext

import (
	"github.com/baetyl/baetyl-go/v2/context"
	mqtt2 "github.com/baetyl/baetyl-go/v2/mqtt"
)

type SystemConfig struct {
	context.SystemConfig `yaml:",inline" json:",inline"`
	Devices              []DeviceInfo `yaml:"devices,omitempty" json:"devices,omitempty"`
}

type DeviceInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	Topic   `yaml:",inline" json:",inline"`
}

type Topic struct {
	Delta       mqtt2.QOSTopic `yaml:"delta,omitempty" json:"delta,omitempty"`
	Report      mqtt2.QOSTopic `yaml:"report,omitempty" json:"report,omitempty"`
	Event       mqtt2.QOSTopic `yaml:"event,omitempty" json:"event,omitempty"`
	Get         mqtt2.QOSTopic `yaml:"get,omitempty" json:"get,omitempty"`
	GetResponse mqtt2.QOSTopic `yaml:"getResponse,omitempty" json:"getResponse,omitempty"`
}
