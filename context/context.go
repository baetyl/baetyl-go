package context

import (
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/pki"
	"github.com/baetyl/baetyl-go/v2/utils"
)

var (
	ErrSystemCertInvalid  = errors.New("system certificate is invalid")
	ErrSystemCertNotFound = errors.New("system certificate is not found")
)

// Context of service
type Context interface {
	// NodeName returns node name from data.
	NodeName() string
	// AppName returns app name from data.
	AppName() string
	// AppVersion returns application version from data.
	AppVersion() string
	// ServiceName returns service name from data.
	ServiceName() string
	// ConfFile returns config file from data.
	ConfFile() string

	// SystemConfig returns the config of baetyl system from data.
	SystemConfig() *SystemConfig

	// Log returns logger interface.
	Log() *log.Logger

	// Wait waits until exit, receiving SIGTERM and SIGINT signals.
	Wait()
	// WaitChan returns wait channel.
	WaitChan() <-chan os.Signal

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

	// CheckSystemCert checks system certificate, if certificate is not found or invalid, returns an error.
	CheckSystemCert() error
	// LoadCustomConfig loads custom config.
	// If 'files' is empty, will load config from default path,
	// else the first file path will be used to load config from.
	LoadCustomConfig(cfg interface{}, files ...string) error
	// NewFunctionHttpClient creates a new function http client.
	NewFunctionHttpClient() (*http.Client, error)
	// NewSystemBrokerClientConfig creates the system config of broker
	NewSystemBrokerClientConfig() (mqtt.ClientConfig, error)
	// NewBrokerClient creates a new broker client.
	NewBrokerClient(mqtt.ClientConfig) (*mqtt.Client, error)
	// NewSystemBrokerClient creates a new system broker client.
	NewSystemBrokerClient([]mqtt.QOSTopic) (*mqtt.Client, error)
}

type ctx struct {
	sync.Map // global cache
	log      *log.Logger
}

// NewContext creates a new context
func NewContext(confFile string) Context {
	if confFile == "" {
		confFile = os.Getenv(KeyConfFile)
	}

	c := &ctx{}
	c.Store(KeyConfFile, confFile)
	c.Store(KeyNodeName, os.Getenv(KeyNodeName))
	c.Store(KeyAppName, os.Getenv(KeyAppName))
	c.Store(KeyAppVersion, os.Getenv(KeyAppVersion))
	c.Store(KeySvcName, os.Getenv(KeySvcName))

	var lfs []log.Field
	if c.NodeName() != "" {
		lfs = append(lfs, log.Any("node", c.NodeName()))
	}
	if c.AppName() != "" {
		lfs = append(lfs, log.Any("app", c.AppName()))
	}
	if c.ServiceName() != "" {
		lfs = append(lfs, log.Any("service", c.ServiceName()))
	}
	c.log = log.With(lfs...)
	c.log.Info("to load config file", log.Any("file", c.ConfFile()))

	sc := &SystemConfig{}
	err := c.LoadCustomConfig(sc)
	if err != nil {
		c.log.Error("failed to load system config, to use default config", log.Error(err))
		utils.UnmarshalYAML(nil, sc)
	}
	// populate configuration
	// if not set in config file, to use value from env.
	// if not set in env, to use default value.
	if sc.Function.Address == "" {
		sc.Function.Address = getFunctionAddress()
	}
	if sc.Function.CA == "" {
		sc.Function.CA = sc.Certificate.CA
	}
	if sc.Function.Key == "" {
		sc.Function.Key = sc.Certificate.Key
	}
	if sc.Function.Cert == "" {
		sc.Function.Cert = sc.Certificate.Cert
	}

	if sc.Broker.Address == "" {
		sc.Broker.Address = getBrokerAddress()
	}
	// auto subscribe link topic for service if service name not nil.
	if sc.Broker.Subscriptions == nil {
		sc.Broker.Subscriptions = []mqtt.QOSTopic{}
	}
	if c.ServiceName() != "" {
		if sc.Broker.ClientID == "" {
			sc.Broker.ClientID = "baetyl-link-" + c.ServiceName()
		}
		sc.Broker.Subscriptions = append(sc.Broker.Subscriptions, mqtt.QOSTopic{QOS: 1, Topic: "$link/" + c.ServiceName()})
	}
	if sc.Broker.CA == "" {
		sc.Broker.CA = sc.Certificate.CA
	}
	if sc.Broker.Key == "" {
		sc.Broker.Key = sc.Certificate.Key
	}
	if sc.Broker.Cert == "" {
		sc.Broker.Cert = sc.Certificate.Cert
	}
	c.Store(KeySysConf, sc)

	_log, err := log.Init(sc.Logger, lfs...)
	if err != nil {
		c.log.Error("failed to init logger", log.Error(err))
	}
	c.log = _log
	c.log.Debug("context is created", log.Any("file", confFile), log.Any("conf", sc))
	return c
}

func (c *ctx) NodeName() string {
	v, ok := c.Load(KeyNodeName)
	if !ok {
		return ""
	}
	return v.(string)
}

func (c *ctx) AppName() string {
	v, ok := c.Load(KeyAppName)
	if !ok {
		return ""
	}
	return v.(string)
}

func (c *ctx) AppVersion() string {
	v, ok := c.Load(KeyAppVersion)
	if !ok {
		return ""
	}
	return v.(string)
}

func (c *ctx) ServiceName() string {
	v, ok := c.Load(KeySvcName)
	if !ok {
		return ""
	}
	return v.(string)
}

func (c *ctx) ConfFile() string {
	v, ok := c.Load(KeyConfFile)
	if !ok {
		return ""
	}
	return v.(string)
}

func (c *ctx) SystemConfig() *SystemConfig {
	v, ok := c.Load(KeySysConf)
	if !ok {
		return nil
	}
	return v.(*SystemConfig)
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

func (c *ctx) CheckSystemCert() error {
	cfg := c.SystemConfig().Certificate
	if !utils.FileExists(cfg.CA) || !utils.FileExists(cfg.Key) || !utils.FileExists(cfg.Cert) {
		return errors.Trace(ErrSystemCertNotFound)
	}
	crt, err := ioutil.ReadFile(cfg.Cert)
	if err != nil {
		return errors.Trace(err)
	}
	info, err := pki.ParseCertificates(crt)
	if err != nil {
		return errors.Trace(err)
	}
	if len(info) != 1 || len(info[0].Subject.OrganizationalUnit) != 1 ||
		info[0].Subject.OrganizationalUnit[0] != KeyBaetyl {
		return errors.Trace(ErrSystemCertInvalid)
	}
	return nil
}

func (c *ctx) LoadCustomConfig(cfg interface{}, files ...string) error {
	f := c.ConfFile()
	if len(files) > 0 && len(files[0]) > 0 {
		f = files[0]
	}
	if utils.FileExists(f) {
		return errors.Trace(utils.LoadYAML(f, cfg))
	}
	return errors.Trace(utils.UnmarshalYAML(nil, cfg))
}

func (c *ctx) NewFunctionHttpClient() (*http.Client, error) {
	err := c.CheckSystemCert()
	if err != nil {
		return nil, errors.Trace(err)
	}
	ops, err := c.SystemConfig().Function.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return http.NewClient(ops), nil
}

func (c *ctx) NewSystemBrokerClientConfig() (mqtt.ClientConfig, error) {
	err := c.CheckSystemCert()
	if err != nil {
		return mqtt.ClientConfig{}, errors.Trace(err)
	}
	config := c.SystemConfig().Broker

	config.Subscriptions = make([]mqtt.QOSTopic, 0)
	copy(config.Subscriptions, c.SystemConfig().Broker.Subscriptions)

	return config, nil
}

func (c *ctx) NewBrokerClient(config mqtt.ClientConfig) (*mqtt.Client, error) {
	ops, err := config.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return mqtt.NewClient(ops), nil
}

func (c *ctx) NewSystemBrokerClient(subTopics []mqtt.QOSTopic) (*mqtt.Client, error) {
	config, err := c.NewSystemBrokerClientConfig()
	if err != nil {
		return nil, err
	}
	if len(subTopics) > 0 {
		config.Subscriptions = append(config.Subscriptions, subTopics...)
	}
	client, err := c.NewBrokerClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
