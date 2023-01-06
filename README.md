baetyl-go
========

[![codecov](https://codecov.io/gh/baetyl/baetyl-go/branch/master/graph/badge.svg)](https://codecov.io/gh/baetyl/baetyl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/baetyl/baetyl-go)](https://goreportcard.com/report/github.com/baetyl/baetyl-go) 
[![License](https://img.shields.io/github/license/baetyl/baetyl-go.svg)](./LICENSE)

# Golang SDK for BAETYL V2

Please use [Old SDK](https://github.com/baetyl/baetyl/tree/stable/1/sdk/baetyl-go) if working on [BAETYL V1](https://github.com/baetyl/baetyl).

# baetyl-go Document
You can use Baetyl SDK to develop moduels with Golang language. It is in baetyl-go warehouse. The functional interface is Context.

At present, the provided SDK capabilities are still relatively limited, and will be gradually strengthened in the future. Other languasges will be supported in future.

## 1. Version
version：git-6a9a8f8

## 2. Basic Function
User developed custom modules can access features provided by baetyl SDK when launch it from the Context defined by baetyl-go.

Calling method：

```go
context.Run(func(ctx context.Context) error
   // business logic
   ......


   return nil
}
```

Context  api：

```go
// Context of service
// Baetyl runtime context
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
```

## 3. Examples
The following takes the implementation of the function module as an example to briefly introduce the usage of baetyl-go.
project address：[baetyl-function](https://github.com/baetyl/baetyl-function/blob/master/cmd/main.go) 

```go
// Config: function 
type Config struct {
	Server http.ServerConfig `yaml:"server" json:"server"`
	Client ClientConfig      `yaml:"client" json:"client"`
}

type ClientConfig struct {
	Grpc GrpcConfig `yaml:"grpc" json:"grpc"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port" json:"port" default:"80"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	Retries int           `yaml:"retries" json:"retries" default:"3"`
}



// program entry
func main() {
    // start service through the context of baetyl-go 
	context.Run(func(ctx context.Context) error {
        // Check if system certificate exists
		if err := ctx.CheckSystemCert(); err != nil {
			return err
		}

		var cfg function.Config
        // load custom config 
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return errors.Trace(err)
		}

        // Create a parser to get the running mode of the current application through context.RunMode() 
		resolver, err := resolve.New(context.RunMode(), ctx)
		if err != nil {
			return errors.Trace(err)
		}
		defer resolver.Close()

        // create and start function http service
		api, err := function.NewAPI(cfg, ctx, resolver)
		if err != nil {
			return errors.Trace(err)
		}
		defer api.Close()


        // wait to exit, listening SIGTERM and SIGINI signal
		ctx.Wait()
		return nil
	})
}
```

function config file

```yaml
server: # server config.Runtimes module that proxies requests to the backend
  address: ":50011" # listening address
  concurrency: # The number of concurrent connections on the server side, if not set, the default value will be used
  disableKeepalive: true # Whether to enable keep-alive connection, the default value is false
  tcpKeepalive: false # Whether to send keep-alive connection, the default value is false
  maxRequestBodySize: # Body max connection number，default value is 4 * 1024 * 1024 Byte
  readTimeout: 1h # read timeout of server connection，default time is unlimited
  writeTimeout: 1h # write timeout of server connection ,default time is unlimited 
  idleTimeout: 1h # when keep alive is started,Under keep alive start condition, the server waits for the idle timeout time of the next message, if the value is 0, multiplex read timeout time
  ca: example/var/lib/baetyl/testcert/ca.crt # Server  CA path
  key: example/var/lib/baetyl/testcert/server.key # Server private key path
  cert: example/var/lib/baetyl/testcert/server.crt # Server public key path

client: # Requests client-side related settings for the backend Runtimes module
  grpc: # Grpc client setting
    port: 80 #  Runtimes port
    timeout: 5m # Request timeout
    retries: 3 # Request retries

logger: # log
  level: info # log level
```

## 4. Other Method
This Chapter introduces the features and function of some usefule packages.

You can find more function api with godoc tool. Refer to Chapter 5 to see how to install and use godoc

### 4.1 context pakage
#### 4.1.1 env
```go
// HostPathLib return HostPathLib
func HostPathLib() (string, error)
// RunMode return run mode of edge.
func RunMode() string
// EdgeNamespace return namespace of edge.
func EdgeNamespace() string
// EdgeSystemNamespace return system namespace of edge.
func EdgeSystemNamespace() string
// BrokerPort return broker port.
func BrokerPort() string
// FunctionPort return http port of function.
func FunctionHttpPort() string
// BrokerHost return broker host.
func BrokerHost() string
// FunctionHost return function host.
func FunctionHost() string
```

#### 4.1.2 platform
```go
// return platform info
// specs.Platform{
//    OS:           runtime.GOOS,
//    Architecture: runtime.GOARCH,
//    // The Variant field will be empty if arch != ARM.
//    Variant: cpuVariant,
// }
func Platform() PlatformInfo
// Returns a string of platform information in the following format
// "%s-%s-%s", pl.OS, pl.Architecture, pl.Variant
func PlatformString() string
```

### 4.2 http package
This package provides http service. You can use the program to complete quick initialization, start http server/client, initiate http request and other functions

#### 4.2.1 client
```go
// NewClient creates a new http client
func NewClient(ops *ClientOptions) *Client


// Call calls the function via HTTP POST
func (c *Client) Call(function string, payload []byte) ([]byte, error)
// PostJSON post data with json content type
func (c *Client) PostJSON(url string, payload []byte, headers ...map[string]string)
// GetJSON get data with json content type
func (c *Client) GetJSON(url string, headers ...map[string]string) ([]byte, error)
func (c *Client) GetURL(url string, header ...map[string]string) 
func (c *Client) PostURL(url string, body io.Reader, header ...map[string]string)
func (c *Client) SendUrl(method, url string, body io.Reader, header ...map[string]string)
```

#### 4.2.2 server
```go
// NewServer new server
func NewServer(cfg ServerConfig, handler fasthttp.RequestHandler) *Server


func (s *Server) Start()
func (s *Server) Close()
```

## 4.3 pki package
Encapsulate goalng's certificate generation and issuance to provide more convenient certificate operation functions

* Support the issuance of root certificates
* Support the issuance of self-signed root certificates
* Support the issuance of sub-certificates
* Support the issuance of sub-certificates with specified private keys

## 4.4 pubsub package
This package provides a memory version of the message queue mechanism that supports publishing and subscription

* Support topic subscription and unsubscription
* Support news release
* Provide an executor that can quickly start a message receiving processor that supports timeout settings

## 4.5 plugin package
This package provides registration mechanism based on the factory mode

## 4.6 tools package
  This package provides a sample program for certificate issuance
  
## 4.7 utils package
This package provides help function

* Certificate parsing
* Configuration parsing
* Default value setting
* Log tracking
* Support zip compression
  
## 4.8 dmcontext package
  Context functions that provide device management functionality 
  
# 5.GoDoc
how to find api document using godoc
  1. Install godoc `go get golang.org/x/tools/cmd/godoc`
  2. make sure godoc is under gosrc path
  3. godoc -http=:6060 The interface is set according to the actual situation
  4. visit following address in broswer `http://0.0.0.0:6060/pkg/`
  5. check baetyl-go document
