package context

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

// Env keys
const (
	EnvKeyConfFile    = "BAETYL_CONF_FILE"
	EnvKeyNodeName    = "BAETYL_NODE_NAME"
	EnvKeyAppName     = "BAETYL_APP_NAME"
	EnvKeyServiceName = "BAETYL_SERVICE_NAME"
	EnvKeyCodePath    = "BAETYL_CODE_PATH"
)

// Context of service
type Context interface {
	// NodeName returns node name.
	NodeName() string
	// AppName returns app name.
	AppName() string
	// ServiceName returns service name.
	ServiceName() string
	// ConfFile returns config file.
	ConfFile() string
	// ServiceConfig returns service config.
	ServiceConfig() ServiceConfig

	// Load returns the value stored in the map for a key, or nil if no value is present.
	// The ok result indicates whether value was found in the map.
	Load(key interface{}) (value interface{}, ok bool)
	// Store sets the value for a key.
	Store(key, value interface{})
	// LoadOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored.
	LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
	// Delete deletes the value for a key.
	Delete(key interface{})

	// LoadCustomConfig loads custom config, if path is empty, will load config from default path.
	LoadCustomConfig(cfg interface{}, files ...string) error
	// Log returns logger interface.
	Log() *log.Logger

	// Wait waits until exit, receiving SIGTERM and SIGINT signals.
	Wait()
	// WaitChan returns wait channel.
	WaitChan() <-chan os.Signal
}

type ctx struct {
	sync.Map
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

	var fs []log.Field
	if c.nodeName != "" {
		fs = append(fs, log.Any("node", c.nodeName))
	}
	if c.appName != "" {
		fs = append(fs, log.Any("app", c.appName))
	}
	if c.serviceName != "" {
		fs = append(fs, log.Any("service", c.serviceName))
	}
	c.log = log.With(fs...)

	err := c.LoadCustomConfig(&c.cfg)
	if err != nil {
		c.log.Error("failed to load service config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, &c.cfg)
	}

	_log, err := log.Init(c.cfg.Logger, fs...)
	if err != nil {
		c.log.Error("failed to init logger", log.Error(err))
	}
	c.log = _log

	if c.cfg.HTTP.Address == "" {
		if c.cfg.HTTP.Key == "" {
			c.cfg.HTTP.Address = "http://baetyl-function:80"
		} else {
			c.cfg.HTTP.Address = "https://baetyl-function:443"
		}
	}

	if c.cfg.MQTT.Address == "" {
		if c.cfg.MQTT.Key == "" {
			c.cfg.MQTT.Address = "tcp://baetyl-broker:1883"
		} else {
			c.cfg.MQTT.Address = "ssl://baetyl-broker:8883"
		}
	}
	c.log.Debug("context is created", log.Any("file", confFile), log.Any("conf", c.cfg))
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

func (c *ctx) ConfFile() string {
	return c.confFile
}

func (c *ctx) ServiceConfig() ServiceConfig {
	return c.cfg
}

func (c *ctx) LoadCustomConfig(cfg interface{}, files ...string) error {
	f := c.confFile
	if len(files) > 0 && len(files[0]) > 0 {
		f = files[0]
	}
	if utils.FileExists(f) {
		return errors.Trace(utils.LoadYAML(f, cfg))
	}
	return errors.Trace(utils.UnmarshalYAML(nil, cfg))
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
