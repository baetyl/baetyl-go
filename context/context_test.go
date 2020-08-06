package context

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func initCert(t *testing.T) string {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	err = os.Setenv(EnvKeyCertPath, dir)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(dir, SystemCertCA), []byte(ca), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(dir, SystemCertCrt), []byte(crt), os.ModePerm)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(dir, SystemCertKey), []byte(key), os.ModePerm)
	assert.NoError(t, err)
	return dir
}

func TestContext(t *testing.T) {
	ctx, err := NewContext("")
	assert.Error(t, err)
	assert.Nil(t, ctx)

	os.Setenv(EnvKeyServiceName, SystemAppInit)
	ctx, err = NewContext("")
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	dir := initCert(t)
	defer os.RemoveAll(dir)

	os.Setenv(EnvKeyConfFile, "file")
	os.Setenv(EnvKeyNodeName, "node")
	os.Setenv(EnvKeyAppName, "app")
	os.Setenv(EnvKeyAppVersion, "v1")
	os.Setenv(EnvKeyServiceName, "service")

	ctx, err = NewContext("")
	assert.NoError(t, err)
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
	assert.Equal(t, "v1", ctx.AppVersion())
	assert.Equal(t, "service", ctx.ServiceName())
	assert.Equal(t, "file", ctx.ConfFile())
	resCa, resCrt, resKey := ctx.GetSystemResource().GetSystemCert()
	assert.Equal(t, ca, string(resCa))
	assert.Equal(t, crt, string(resCrt))
	assert.Equal(t, key, string(resKey))
	cfg := ctx.ServiceConfig()
	assert.Equal(t, "http://baetyl-function:80", cfg.HTTP.Address)
	assert.Equal(t, "tcp://baetyl-broker:1883", cfg.MQTT.Address)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, "json", cfg.Logger.Encoding)
	assert.Empty(t, cfg.Logger.Filename)
	assert.False(t, cfg.Logger.Compress)
	assert.Equal(t, 15, cfg.Logger.MaxAge)
	assert.Equal(t, 50, cfg.Logger.MaxSize)
	assert.Equal(t, 15, cfg.Logger.MaxBackups)

	ctx, err = NewContext("../example/etc/baetyl/service.yml")
	assert.NoError(t, err)
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
	assert.Equal(t, "v1", ctx.AppVersion())
	assert.Equal(t, "service", ctx.ServiceName())
	assert.Equal(t, "../example/etc/baetyl/service.yml", ctx.ConfFile())
	cfg = ctx.ServiceConfig()
	assert.Equal(t, "https://baetyl-function:443", cfg.HTTP.Address)
	assert.Equal(t, "ssl://baetyl-broker:8883", cfg.MQTT.Address)
	assert.Equal(t, "debug", cfg.Logger.Level)
	assert.Equal(t, "console", cfg.Logger.Encoding)
	assert.Empty(t, cfg.Logger.Filename)
	assert.False(t, cfg.Logger.Compress)
	assert.Equal(t, 15, cfg.Logger.MaxAge)
	assert.Equal(t, 50, cfg.Logger.MaxSize)
	assert.Equal(t, 15, cfg.Logger.MaxBackups)
}
