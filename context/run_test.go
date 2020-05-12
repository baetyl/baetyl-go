package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	Run(func(ctx Context) error {
		assert.Equal(t, "etc/baetyl/service.yml", ctx.ConfFile())

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
		return nil
	})
}
