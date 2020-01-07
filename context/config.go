package context

import (
	"github.com/baetyl/baetyl-go/kv"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
	"google.golang.org/grpc"
)

// ServiceConfig base config of service
type ServiceConfig struct {
	Mqtt   mqtt.ClientConfig `yaml:"mqtt" json:"mqtt"`
	Link   link.ClientConfig `yaml:"link" json:"link"`
	Logger log.Config        `yaml:"logger" json:"logger"`
}

// Client a grpc client that connects to a grpc server
type Client struct {
	conf link.ClientConfig
	conn *grpc.ClientConn
	KV   kv.KVServiceClient
}
