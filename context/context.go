package context

import (
	"fmt"
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

// All keys
const (
	KeyBaetyl              = "BAETYL"
	KeyConfFile            = "BAETYL_CONF_FILE"
	KeyNodeName            = "BAETYL_NODE_NAME"
	KeyAppName             = "BAETYL_APP_NAME"
	KeyAppVersion          = "BAETYL_APP_VERSION"
	KeySvcName             = "BAETYL_SERVICE_NAME"
	KeySysConf             = "BAETYL_SYSTEM_CONF"
	KeyRunMode             = "BAETYL_RUN_MODE"
	KeyBrokerHost          = "BAETYL_BROKER_HOST"
	KeyBrokerPort          = "BAETYL_BROKER_PORT"
	KeyFunctionHost        = "BAETYL_FUNCTION_HOST"
	KeyFunctionHttpPort    = "BAETYL_FUNCTION_HTTP_PORT"
	KeyEdgeNamespace       = "BAETYL_EDGE_NAMESPACE"
	KeyEdgeSystemNamespace = "BAETYL_EDGE_SYSTEM_NAMESPACE"

	BaetylEdgeNamespace          = "baetyl-edge"
	BaetylEdgeSystemNamespace    = "baetyl-edge-system"
	BaetylBrokerSystemPort       = "50010"
	BaetylFunctionSystemHttpPort = "50011"
	BaetylFunctionSystemGrpcPort = "50012"

	RunModeKube   = "kube"
	RunModeNative = "native"
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
	// RunMode return run mode.
	RunMode() string
	// BrokerHost return broker host.
	BrokerHost() string
	// BrokerPort return broker port.
	BrokerPort() string
	// FunctionHost return function host.
	FunctionHost() string
	// FunctionHttpPort return http port of function.
	FunctionHttpPort() string
	// EdgeNamespace return namespace of edge.
	EdgeNamespace() string
	// EdgeSystemNamespace return system namespace of edge.
	EdgeSystemNamespace() string
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
	c.Store(KeyRunMode, os.Getenv(KeyRunMode))

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
		sc.Function.Address = c.getFunctionAddress()
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
		sc.Broker.Address = c.getBrokerAddress()
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

// RunMode return run mode.
func (c *ctx) RunMode() string {
	v, ok := c.Load(KeyRunMode)
	if !ok {
		return RunModeKube
	}
	return v.(string)
}

// BrokerHost return broker host.
func (c *ctx) BrokerHost() string {
	if host := os.Getenv(KeyBrokerHost); host != "" {
		return host
	}

	if c.RunMode() == RunModeNative {
		return "127.0.0.1"
	}
	return fmt.Sprintf("%s.%s", "baetyl-broker", BaetylEdgeNamespace)
}

// BrokerPort return broker port.
func (c *ctx) BrokerPort() string {
	if port := os.Getenv(KeyBrokerPort); port != "" {
		return port
	}
	return BaetylBrokerSystemPort
}

// FunctionHost return function host.
func (c *ctx) FunctionHost() string {
	if host := os.Getenv(KeyFunctionHost); host != "" {
		return host
	}

	if c.RunMode() == RunModeNative {
		return "127.0.0.1"
	}
	return fmt.Sprintf("%s.%s", "baetyl-function", BaetylEdgeSystemNamespace)
}

// FunctionPort return http port of function.
func (c *ctx) FunctionHttpPort() string {
	if port := os.Getenv(KeyFunctionHttpPort); port != "" {
		return port
	}
	return BaetylFunctionSystemHttpPort
}

// EdgeNamespace return namespace of edge.
func (c *ctx) EdgeNamespace() string {
	if port := os.Getenv(KeyEdgeNamespace); port != "" {
		return port
	}
	return BaetylEdgeNamespace
}

// EdgeSystemNamespace return system namespace of edge.
func (c *ctx) EdgeSystemNamespace() string {
	if port := os.Getenv(KeyEdgeSystemNamespace); port != "" {
		return port
	}
	return BaetylEdgeSystemNamespace
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

func (c *ctx) getBrokerAddress() string {
	return fmt.Sprintf("%s://%s:%s", "ssl", c.BrokerHost(), c.BrokerPort())
}

func (c *ctx) getFunctionAddress() string {
	return fmt.Sprintf("%s://%s:%s", "https", c.FunctionHost(), c.FunctionHttpPort())
}
