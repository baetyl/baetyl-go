package context

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
	"github.com/baetyl/baetyl-go/utils"
)

// Env keys
const (
	EnvKeyNodeName    = "BAETYL_NODE_NAME"
	EnvKeyAppName     = "BAETYL_APP_NAME"
	EnvKeyServiceName = "BAETYL_SERVICE_NAME"
)

const (
	// DefaultConfFile service config path by default
	DefaultConfFile = "/etc/baetyl/service.yml"
	// DefaultFunctionAddress middleware function address by default
	DefaultFunctionAddress = "https://baetyl-function:8880"
	// DefaultBrokerMqttAddress middleware broker mqtt address by default
	DefaultBrokerMqttAddress = "ssl://baetyl-broker:8883"
	// DefaultBrokerLinkAddress middleware broker link address by default
	DefaultBrokerLinkAddress = "ssl://baetyl-broker:8886"
)

// Context of service
type Context interface {
	// creates a MQTT Client that connects to the broker through system configuration
	NewMQTTClient(string, mqtt.Observer, []mqtt.QOSTopic) (*mqtt.Client, error)
	// creates a Link Client that connects to the broker through system configuration
	NewLinkClient(link.Observer) (*link.Client, error)
	// returns logger interface
	Log() *log.Logger
	// waiting to exit, receiving SIGTERM and SIGINT signals
	Wait()
	// returns wait channel
	WaitChan() <-chan os.Signal
}

type ctx struct {
	nn  string
	an  string
	sn  string
	cfg ServiceConfig
	log *log.Logger
}

func newContext() *ctx {
	nn := os.Getenv(EnvKeyNodeName)
	an := os.Getenv(EnvKeyAppName)
	sn := os.Getenv(EnvKeyServiceName)
	fs := []log.Field{log.Any("node", nn), log.Any("app", an), log.Any("service", sn)}
	l := log.With(fs...)

	var err error
	var cfg ServiceConfig
	if utils.FileExists(DefaultConfFile) {
		err = utils.LoadYAML(DefaultConfFile, &cfg)
	} else {
		err = utils.UnmarshalYAML(nil, &cfg)
	}
	if err != nil {
		l.Error("failed to load config", log.Error(err))
	}
	l, err = log.Init(cfg.Logger, fs...)
	if err != nil {
		l.Error("failed to init logger", log.Error(err))
	}
	if cfg.Mqtt.Address == "" {
		cfg.Mqtt.Address = DefaultBrokerMqttAddress
	}
	if cfg.Link.Address == "" {
		cfg.Link.Address = DefaultBrokerLinkAddress
	}
	c := &ctx{
		nn:  nn,
		an:  an,
		sn:  sn,
		cfg: cfg,
		log: l,
	}
	l.Info("context is created", log.Any("config", cfg))
	return c
}

func (c *ctx) NewMQTTClient(cid string, obs mqtt.Observer, topics []mqtt.QOSTopic) (*mqtt.Client, error) {
	cc := c.cfg.Mqtt
	if cid != "" {
		cc.ClientID = cid
	}
	cli, err := mqtt.NewClient(cc, obs)
	if err != nil {
		return nil, err
	}
	var subs []mqtt.Subscription
	for _, topic := range topics {
		subs = append(subs, mqtt.Subscription{Topic: topic.Topic, QOS: mqtt.QOS(topic.QOS)})
	}
	if len(subs) > 0 {
		err = cli.Subscribe(subs)
		if err != nil {
			return nil, err
		}
	}
	return cli, nil
}

func (c *ctx) NewLinkClient(obs link.Observer) (*link.Client, error) {
	cc := c.cfg.Link
	cli, err := link.NewClient(cc, obs)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (c *ctx) LoadConfig(cfg interface{}) error {
	return utils.LoadYAML(DefaultConfFile, cfg)
}

func (c *ctx) NodeName() string {
	return c.nn
}

func (c *ctx) AppName() string {
	return c.an
}

func (c *ctx) ServiceName() string {
	return c.sn
}

func (c *ctx) Config() ServiceConfig {
	return c.cfg
}

func (c *ctx) Log() *log.Logger {
	return c.log
}

func (c *ctx) Wait() {
	<-c.WaitChan()
}

func (c *ctx) WaitChan() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	return sig
}
