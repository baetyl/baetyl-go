package context

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baetyl/baetyl-go/api"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
	"github.com/baetyl/baetyl-go/utils"
)

// Mode keys
const (
	ModeNative = "native"
	ModeDocker = "docker"
)

// Env keys
const (
	// new envs
	EnvKeyHostID              = "BAETYL_HOST_ID"
	EnvKeyHostOS              = "BAETYL_HOST_OS"
	EnvKeyHostSN              = "BAETYL_HOST_SN"
	EnvKeyAPISocket           = "BAETYL_API_SOCKET"
	EnvKeyAPIAddress          = "BAETYL_API_ADDRESS"
	EnvKeyServiceMode         = "BAETYL_SERVICE_MODE"
	EnvKeyServiceName         = "BAETYL_SERVICE_NAME"
	EnvKeyServiceToken        = "BAETYL_SERVICE_TOKEN"
	EnvKeyServiceInstanceName = "BAETYL_SERVICE_INSTANCE_NAME"
)

// Path keys
const (
	// AppConfFileName application config file name
	AppConfFileName = "application.yml"
	// AppBackupFileName application backup configuration file
	AppBackupFileName = "application.yml.old"
	// AppStatsFileName application stats file name
	AppStatsFileName = "application.stats"
	// MetadataFileName application metadata file name
	MetadataFileName = "metadata.yml"

	// BinFile the file path of master binary
	DefaultBinFile = "bin/baetyl"
	// DefaultBinBackupFile the backup file path of master binary
	DefaultBinBackupFile = "bin/baetyl.old"
	// DefaultSockFile sock file of baetyl by default
	DefaultSockFile = "var/run/baetyl.sock"
	// DefaultConfFile config path of the service by default
	DefaultConfFile = "etc/baetyl/service.yml"
	// DefaultDBDir db dir of the service by default
	DefaultDBDir = "var/db/baetyl"
	// DefaultRunDir  run dir of the service by default
	DefaultRunDir = "var/run/baetyl"
	// DefaultLogDir  log dir of the service by default
	DefaultLogDir = "var/log/baetyl"
)

// Context of service
type Context interface {
	// returns the system configuration of the service, such as hub and logger
	Config() *ServiceConfig
	// loads the custom configuration of the service
	LoadConfig(interface{}) error
	// creates a MQTT Client that connects to the broker through system configuration
	NewMQTTClient(string, mqtt.Observer, []mqtt.QOSTopic) (*mqtt.Client, error)
	// creates a Link Client that connects to the broker through system configuration
	NewLinkClient(link.Observer) (*link.Client, error)
	// returns logger interface
	Log() *log.Logger
	// check running mode
	IsNative() bool
	// waiting to exit, receiving SIGTERM and SIGINT signals
	Wait()
	// returns wait channel
	WaitChan() <-chan os.Signal

	// Master KV API

	// set kv
	SetKV(kv api.KV) error
	// set kv which supports context
	SetKVConext(ctx context.Context, kv api.KV) error
	// get kv
	GetKV(k []byte) (*api.KV, error)
	// get kv which supports context
	GetKVConext(ctx context.Context, k []byte) (*api.KV, error)
	// del kv
	DelKV(k []byte) error
	// del kv which supports context
	DelKVConext(ctx context.Context, k []byte) error
	// list kv with prefix
	ListKV(p []byte) ([]*api.KV, error)
	// list kv with prefix which supports context
	ListKVContext(ctx context.Context, p []byte) ([]*api.KV, error)
}

type ctx struct {
	sn  string // service name
	in  string // instance name
	md  string // running mode
	cfg ServiceConfig
	log *log.Logger
	*Client
}

func newContext() (*ctx, error) {
	var cfg ServiceConfig
	md := os.Getenv(EnvKeyServiceMode)
	sn := os.Getenv(EnvKeyServiceName)
	in := os.Getenv(EnvKeyServiceInstanceName)

	err := utils.LoadYAML(DefaultConfFile, &cfg)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to load config: %s\n", sn, in, err.Error())
	}
	logger, err := log.Init(cfg.Logger, log.Any("service", sn), log.Any("instance", in))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to init logger: %s\n", sn, in, err.Error())
		logger = log.With(log.Any("service", sn), log.Any("instance", in))
		logger.Error("failed to init logger", log.Error(err))
	}
	cli, err := NewEnvClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to create master client: %s\n", sn, in, err.Error())
		logger.Error("failed to create master client", log.Error(err))
	}
	return &ctx{
		sn:     sn,
		in:     in,
		md:     md,
		cfg:    cfg,
		log:    logger,
		Client: cli,
	}, nil
}

func (c *ctx) NewMQTTClient(cid string, obs mqtt.Observer, topics []mqtt.QOSTopic) (*mqtt.Client, error) {
	if c.cfg.Mqtt.Address == "" {
		return nil, fmt.Errorf("mqtt endpoint not configured")
	}
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
	if c.cfg.Link.Address == "" {
		return nil, fmt.Errorf("link endpoint not configured")
	}
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

func (c *ctx) Config() *ServiceConfig {
	return &c.cfg
}

func (c *ctx) Log() *log.Logger {
	return c.log
}

func (c *ctx) Wait() {
	<-c.WaitChan()
	c.Close()
}

func (c *ctx) IsNative() bool {
	return c.md == ModeNative
}

func (c *ctx) WaitChan() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	return sig
}
