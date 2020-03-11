package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	os.Setenv(EnvKeyNodeName, "node")
	os.Setenv(EnvKeyAppName, "app")
	os.Setenv(EnvKeyServiceName, "service")

	ctx := newContext()
	assert.Equal(t, "node", ctx.NodeName())
	assert.Equal(t, "app", ctx.AppName())
	assert.Equal(t, "service", ctx.ServiceName())
	cfg := ctx.Config()
	assert.Equal(t, "ssl://baetyl-broker:8883", cfg.Mqtt.Address)
	assert.Equal(t, "baetyl-broker:8886", cfg.Link.Address)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, "json", cfg.Logger.Encoding)
	assert.Empty(t, cfg.Logger.Filename)
	assert.False(t, cfg.Logger.Compress)
	assert.Equal(t, 15, cfg.Logger.MaxAge)
	assert.Equal(t, 50, cfg.Logger.MaxSize)
	assert.Equal(t, 15, cfg.Logger.MaxBackups)
}
