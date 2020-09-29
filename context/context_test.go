package context

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/utils"
)

func TestContext(t *testing.T) {
	os.Setenv(KeyRunMode, "")
	expected := &SystemConfig{
		Certificate: utils.Certificate{
			CA:                 "var/lib/baetyl/system/certs/ca.pem",
			Key:                "var/lib/baetyl/system/certs/key.pem",
			Cert:               "var/lib/baetyl/system/certs/crt.pem",
			InsecureSkipVerify: false,
			ClientAuthType:     0,
		},
		Function: http.ClientConfig{
			Address:               "https://baetyl-function.baetyl-edge-system:" + baetylFunctionSystemHttpPort,
			Timeout:               30000000000,
			KeepAlive:             30000000000,
			MaxIdleConns:          100,
			IdleConnTimeout:       90000000000,
			TLSHandshakeTimeout:   10000000000,
			ExpectContinueTimeout: 1000000000,
			Certificate: utils.Certificate{
				CA:                 "var/lib/baetyl/system/certs/ca.pem",
				Key:                "var/lib/baetyl/system/certs/key.pem",
				Cert:               "var/lib/baetyl/system/certs/crt.pem",
				InsecureSkipVerify: false,
				ClientAuthType:     0,
			},
		},
		Broker: mqtt.ClientConfig{
			Address:              "ssl://baetyl-broker.baetyl-edge:" + baetylBrokerSystemPort,
			Username:             "",
			Password:             "",
			ClientID:             "",
			CleanSession:         false,
			Timeout:              30000000000,
			KeepAlive:            30000000000,
			MaxReconnectInterval: 180000000000,
			MaxCacheMessages:     10,
			DisableAutoAck:       false,
			Subscriptions:        []mqtt.QOSTopic{},
			Certificate: utils.Certificate{
				CA:                 "var/lib/baetyl/system/certs/ca.pem",
				Key:                "var/lib/baetyl/system/certs/key.pem",
				Cert:               "var/lib/baetyl/system/certs/crt.pem",
				InsecureSkipVerify: false,
				ClientAuthType:     0,
			},
		},
		Logger: log.Config{
			Level:       "info",
			Encoding:    "json",
			Filename:    "",
			Compress:    false,
			MaxAge:      15,
			MaxSize:     50,
			MaxBackups:  15,
			EncodeTime:  "",
			EncodeLevel: "",
		},
	}

	ctx := NewContext("")
	assert.Equal(t, "", ctx.NodeName())
	assert.Equal(t, "", ctx.AppName())
	assert.Equal(t, "", ctx.AppVersion())
	assert.Equal(t, "", ctx.ServiceName())
	assert.Equal(t, "", ctx.ConfFile())
	assert.Equal(t, expected, ctx.SystemConfig())

	os.Setenv(KeyConfFile, "file")
	os.Setenv(KeyNodeName, "node")
	os.Setenv(KeyAppName, "app")
	os.Setenv(KeyAppVersion, "v1")
	os.Setenv(KeySvcName, "service")
	ctx = NewContext("")
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
	assert.Equal(t, "v1", ctx.AppVersion())
	assert.Equal(t, "service", ctx.ServiceName())
	assert.Equal(t, "file", ctx.ConfFile())
	expected.Broker.ClientID = "baetyl-link-service"
	expected.Broker.Subscriptions = append(expected.Broker.Subscriptions, mqtt.QOSTopic{QOS: 1, Topic: "$link/service"})
	assert.Equal(t, expected, ctx.SystemConfig())

	ctx = NewContext("../example/etc/baetyl/service.yml")
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
	assert.Equal(t, "v1", ctx.AppVersion())
	assert.Equal(t, "service", ctx.ServiceName())
	assert.Equal(t, "../example/etc/baetyl/service.yml", ctx.ConfFile())
	expected.Certificate.CA = "example/var/lib/baetyl/testcert/ca.pem"
	expected.Certificate.Key = "example/var/lib/baetyl/testcert/client.key"
	expected.Certificate.Cert = "example/var/lib/baetyl/testcert/client.pem"
	expected.Function.Address = "https://baetyl-function:8880"
	expected.Function.CA = expected.Certificate.CA
	expected.Function.Key = expected.Certificate.Key
	expected.Function.Cert = expected.Certificate.Cert
	expected.Broker.Address = "ssl://baetyl-broker:8883"
	expected.Broker.CA = expected.Certificate.CA
	expected.Broker.Key = expected.Certificate.Key
	expected.Broker.Cert = expected.Certificate.Cert
	expected.Logger.Filename = "var/log/service.log"
	expected.Logger.Level = "debug"
	expected.Logger.Encoding = "console"
	assert.Equal(t, expected, ctx.SystemConfig())

	fc, err := ctx.NewFunctionHttpClient()
	assert.EqualError(t, err, ErrSystemCertNotFound.Error())
	assert.Nil(t, fc)

	_, err = ctx.NewSystemBrokerClientConfig()
	assert.EqualError(t, err, ErrSystemCertNotFound.Error())
}

func TestContext_CheckSystemCert(t *testing.T) {
	dir := initCert(t)
	defer os.RemoveAll(dir)
	ctx := NewContext("")
	err := ctx.CheckSystemCert()
	assert.NoError(t, err)

	fc, err := ctx.NewFunctionHttpClient()
	assert.NoError(t, err)
	assert.NotNil(t, fc)

	config, err := ctx.NewSystemBrokerClientConfig()
	assert.NoError(t, err)

	config.Subscriptions = append(config.Subscriptions, mqtt.QOSTopic{
		QOS:   0,
		Topic: "test",
	})

	config2, err := ctx.NewSystemBrokerClientConfig()
	assert.NoError(t, err)

	assert.NotEqual(t, config, config2)

	bc, err := ctx.NewBrokerClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, bc)
}

func initCert(t *testing.T) string {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	err = os.Chdir(dir)
	assert.NoError(t, err)

	var cfg SystemConfig
	err = utils.UnmarshalYAML(nil, &cfg)
	assert.NoError(t, err)
	fmt.Println(cfg)

	err = os.MkdirAll(filepath.Dir(cfg.Certificate.CA), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(cfg.Certificate.CA, []byte(ca), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(cfg.Certificate.Cert, []byte(crt), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(cfg.Certificate.Key, []byte(key), os.ModePerm)
	assert.NoError(t, err)
	return dir
}

const (
	ca = `-----BEGIN CERTIFICATE-----
MIICcDCCAhagAwIBAgIDAYagMAoGCCqGSM49BAMCMIGsMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEZMBcGA1UEBxMQSGFpZGlhbiBEaXN0cmljdDEVMBMG
A1UECRMMQmFpZHUgQ2FtcHVzMQ8wDQYDVQQREwYxMDAwOTMxHjAcBgNVBAoTFUxp
bnV4IEZvdW5kYXRpb24gRWRnZTEPMA0GA1UECxMGQkFFVFlMMRcwFQYDVQQDEw5j
bGllbnQucm9vdC5jYTAgFw0yMDAzMjcwOTU1MzVaGA8yMDUwMDMyNzA5NTUzNVow
gawxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlk
aWFuIERpc3RyaWN0MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEw
MDA5MzEeMBwGA1UEChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZC
QUVUWUwxFzAVBgNVBAMTDmNsaWVudC5yb290LmNhMFkwEwYHKoZIzj0CAQYIKoZI
zj0DAQcDQgAE6FKiZEEPgkBbQwqd8vHX7+NPEa3j33WbTz0/xMcoAZTbg9yMZR/9
BfNjArw3rUHc4K9aMNnlFYXu1RCgdLiIxqMjMCEwDgYDVR0PAQH/BAQDAgGGMA8G
A1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0EAwIDSAAwRQIhAKV27nEL++GfZoA8WGBw
q2MjYpZ2G6tvqQIa9oBI7z/dAiA3z47euYsWN2/rPjRoHTDAa9aZ6bBOXr8t0fKX
sgQJMw==
-----END CERTIFICATE-----
`
	crt = `-----BEGIN CERTIFICATE-----
MIICnDCCAkKgAwIBAgIIFiY2JF4KpRAwCgYIKoZIzj0EAwIwgaUxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxEDAOBgNV
BAMTB3Jvb3QuY2EwHhcNMjAwNzI5MTEzNzI3WhcNNDAwNzI0MTEzNzI3WjCBpDEL
MAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxGTAXBgNVBAcTEEhhaWRpYW4g
RGlzdHJpY3QxFTATBgNVBAkTDEJhaWR1IENhbXB1czEPMA0GA1UEERMGMTAwMDkz
MR4wHAYDVQQKExVMaW51eCBGb3VuZGF0aW9uIEVkZ2UxDzANBgNVBAsTBkJBRVRZ
TDEPMA0GA1UEAxMGc2VydmVyMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEY7n2
J/ZS1eA8WuJJlYYrNqI5gBI0S4dDxuU33+pRaiM6+i1PdhHqIQvJ2Bn2NdpZ8d7p
9Jz3ESwJjxOp3irb+KNbMFkwDgYDVR0PAQH/BAQDAgWgMA8GA1UdJQQIMAYGBFUd
JQAwDAYDVR0TAQH/BAIwADAoBgNVHREEITAfhwQAAAAAhwR/AAABhhFodHRwczov
L2xvY2FsaG9zdDAKBggqhkjOPQQDAgNIADBFAiEArXD3gq8jSIrZPph+0s1x+ETD
sm6msICrv3/jDHNkafcCIAKBczGOgz19dN8ySrbtq6kfJOkNVBJuPD3urTM+wQii
-----END CERTIFICATE-----
`
	key = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIDr9pIYtROzs4tVZbaY5osMKnSwJyoSPaYHTQty2obFYoAoGCCqGSM49
AwEHoUQDQgAEY7n2J/ZS1eA8WuJJlYYrNqI5gBI0S4dDxuU33+pRaiM6+i1PdhHq
IQvJ2Bn2NdpZ8d7p9Jz3ESwJjxOp3irb+A==
-----END EC PRIVATE KEY-----
`
)
