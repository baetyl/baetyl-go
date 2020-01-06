package context

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestContext(t *testing.T) {
	os.Setenv(EnvKeyServiceMode, "docker")
	os.Setenv(EnvKeyServiceName, "baetyl")
	os.Setenv(EnvKeyServiceInstanceName, "baetyl")

	ctx, err := newContext()
	assert.NoError(t, err)
	assert.Equal(t, ctx.sn, os.Getenv(EnvKeyServiceName))
	assert.Equal(t, ctx.in, os.Getenv(EnvKeyServiceInstanceName))
	assert.Equal(t, ctx.md, os.Getenv(EnvKeyServiceMode))
	assert.Equal(t, ctx.log, log.With())

	var loggerConf log.Config
	ctx.cfg = ServiceConfig{
		Mqtt: mqtt.ClientConfig{
			Address: "tcp://0.0.0.0:51080",
		},
		Link: link.ClientConfig{
			Address: "http://127.0.0.1:51090",
		},
		Logger: loggerConf,
	}
	cid := "baetyl"
	obs := new(mockMqttObserver)
	cli, err := ctx.NewMQTTClient(cid, obs, []mqtt.QOSTopic{})
	assert.NoError(t, err)
	cli.Close()

	cli2, err := ctx.NewLinkClient(new(mockLinkObserver))
	assert.NoError(t, err)
	cli2.Close()

	err = ctx.LoadConfig(ServiceConfig{})
	assert.Error(t, err)
	assert.Equal(t, "open etc/baetyl/service.yml: no such file or directory", err.Error())

	conf := ctx.Config()
	assert.Equal(t, &ctx.cfg, conf)

	logger := ctx.Log()
	assert.Equal(t, ctx.log, logger)

	native := ctx.IsNative()
	assert.Equal(t, false, native)
}

type mockMqttObserver struct{}

func (*mockMqttObserver) OnPublish(*packet.Publish) error {
	return nil
}

func (*mockMqttObserver) OnPuback(*packet.Puback) error {
	return nil
}

func (*mockMqttObserver) OnError(err error) {}

type mockLinkObserver struct{}

func (o *mockLinkObserver) OnMsg(*link.Message) error {
	return nil
}

func (o *mockLinkObserver) OnAck(msg *link.Message) error {
	return nil
}

func (o *mockLinkObserver) OnErr(err error) {
	return
}
