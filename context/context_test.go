package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	os.Setenv(EnvKeyConfFile, "file")
	os.Setenv(EnvKeyNodeName, "node")
	os.Setenv(EnvKeyAppName, "app")
	os.Setenv(EnvKeyServiceName, "service")

	ctx := NewContext("")
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
	assert.Equal(t, "service", ctx.ServiceName())
	assert.Equal(t, "file", ctx.ConfFile())
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

	ctx = NewContext("../example/etc/baetyl/service.yml")
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
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
