package context

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

// Env keys
const (
	EnvKeyConfFile    = "BAETYL_CONF_FILE"
	EnvKeyNodeName    = "BAETYL_NODE_NAME"
	EnvKeyAppName     = "BAETYL_APP_NAME"
	EnvKeyServiceName = "BAETYL_SERVICE_NAME"
)

// Context of service
type Context interface {
	// returns node name
	NodeName() string
	// returns app name
	AppName() string
	// returns service name
	ServiceName() string
	// returns service config
	ServiceConfig() ServiceConfig
	// leads custom config, if path is empty, will load config from default path
	LoadCustomConfig(cfg interface{}, path string) error
	// returns logger interface
	Log() *log.Logger
	// waiting to exit, receiving SIGTERM and SIGINT signals
	Wait()
	// returns wait channel
	WaitChan() <-chan os.Signal
}

type ctx struct {
	cfg ServiceConfig
	log *log.Logger

	nodeName    string
	appName     string
	serviceName string
	confFile    string
	httpAddress string
	mqttAddress string
	linkAddress string
}

// NewContext creates a new context
func NewContext(confFile string) Context {
	if confFile == "" {
		confFile = os.Getenv(EnvKeyConfFile)
	}
	c := &ctx{
		confFile:    confFile,
		nodeName:    os.Getenv(EnvKeyNodeName),
		appName:     os.Getenv(EnvKeyAppName),
		serviceName: os.Getenv(EnvKeyServiceName),
	}

	fs := []log.Field{log.Any("node", c.nodeName), log.Any("app", c.appName), log.Any("service", c.serviceName)}
	c.log = log.With(fs...)

	var err error
	if utils.FileExists(c.confFile) {
		err = utils.LoadYAML(c.confFile, &c.cfg)
	} else {
		err = utils.UnmarshalYAML(nil, &c.cfg)
	}
	if err != nil {
		c.log.Error("failed to load service config", log.Error(err))
	}
	c.log, err = log.Init(c.cfg.Logger, fs...)
	if err != nil {
		c.log.Error("failed to init logger", log.Error(err))
	}
	if c.cfg.HTTP.Address == "" {
		if c.cfg.HTTP.Key == "" {
			c.cfg.HTTP.Address = "http://baetyl-function:8880"
		} else {
			c.cfg.HTTP.Address = "https://baetyl-function:8880"
		}
	}
	if c.cfg.MQTT.Address == "" {
		if c.cfg.MQTT.Key == "" {
			c.cfg.MQTT.Address = "tcp://baetyl-broker:1883"
		} else {
			c.cfg.MQTT.Address = "ssl://baetyl-broker:8883"
		}
	}
	if c.cfg.Link.Address == "" {
		c.cfg.Link.Address = "link://baetyl-broker:8886"
	}
	c.log.Info("context is created", log.Any("config", c.cfg))
	return c
}

func (c *ctx) NodeName() string {
	return c.nodeName
}

func (c *ctx) AppName() string {
	return c.appName
}

func (c *ctx) ServiceName() string {
	return c.serviceName
}

func (c *ctx) ServiceConfig() ServiceConfig {
	return c.cfg
}

func (c *ctx) LoadCustomConfig(cfg interface{}, path string) error {
	if path == "" {
		path = c.confFile
	}
	return utils.LoadYAML(path, cfg)
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
