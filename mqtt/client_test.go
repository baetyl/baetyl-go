package mqtt

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/flow"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestMqttClientConnectErrorMissingAddress(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	obs := newMockObserver(t)
	cli, err := NewClient(ClientConfig{}, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs.assertErrs(errors.New("parse : empty url"))
}

func TestMqttClientConnectErrorWrongPort(t *testing.T) {
	cc := newConfig("1234567")
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs.assertErrs(errors.New("dial tcp: address 1234567: invalid port"))
}

func TestMqttClientConnectWithCredentials(t *testing.T) {
	connect := connectPacket()
	connect.Username = "test"
	connect.Password = "test"

	connack := connackPacket()
	connack.ReturnCode = BadUsernameOrPassword

	broker := flow.New().Debug().
		Receive(connect).
		Send(connack).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Username = "test"
	cc.Password = "test"
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs.assertErrs(errors.New("connection refused: bad user name or password"))
	safeReceive(done)
}

func TestMqttClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = NotAuthorized

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs.assertErrs(errors.New("connection refused: not authorized"))
	safeReceive(done)
}

func TestMqttClientExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(NewPingresp()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs.assertErrs(ErrClientExpectedConnack)
	safeReceive(done)
}

func TestMqttClientNotExpectedConnack(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Send(connackPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	obs.assertErrs(ErrClientAlreadyConnecting)
	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientKeepAlive(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	connect := connectPacket()
	connect.KeepAlive = 1

	pingreq := NewPingreq()
	pingresp := NewPingresp()
	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		Send(pingresp).
		Receive(pingreq).
		Send(pingresp).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.KeepAlive = time.Millisecond * 100
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	time.Sleep(250 * time.Millisecond)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientKeepAliveTimeout(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 1

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(NewPingreq()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.KeepAlive = time.Millisecond * 100
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	safeReceive(done)
	obs.assertErrs(ErrClientMissingPong)
	assert.NoError(t, cli.Close())
}

func TestMqttClientKeepAliveNone(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	<-time.After(time.Second)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := NewSuback()
	suback.ReturnCodes = []QOS{0}
	suback.ID = 1

	publish := NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	err = cli.Subscribe([]Subscription{Subscription{Topic: "test"}})
	assert.NoError(t, err)

	err = cli.Publish(publish.Message.QOS, publish.Message.Topic, publish.Message.Payload, publish.ID, publish.Message.Retain, publish.Dup)
	assert.NoError(t, err)
	obs.assertPkts(publish)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientPublishSubscribeQOS1(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test", QOS: 1}}
	subscribe.ID = 1

	suback := NewSuback()
	suback.ReturnCodes = []QOS{1}
	suback.ID = 1

	publish := NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.ID = 2

	puback := NewPuback()
	puback.ID = 2

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(puback).
		Send(publish).
		Receive(puback).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	err = cli.Subscribe([]Subscription{Subscription{Topic: "test", QOS: 1}})
	assert.NoError(t, err)

	err = cli.Publish(publish.Message.QOS, publish.Message.Topic, publish.Message.Payload, publish.ID, publish.Message.Retain, publish.Dup)
	assert.NoError(t, err)

	obs.assertPkts(puback, publish)

	err = cli.Send(puback)
	assert.NoError(t, err)
	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientUnexpectedClose(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	safeReceive(done)
	obs.assertErrs(io.EOF)
	assert.NoError(t, cli.Close())
}

func TestMqttClientConnackTimeout1(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	obs.assertErrs(io.EOF)
	safeReceive(done)
	cli.Close()
}

func TestMqttClientConnackTimeout2(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Timeout = time.Millisecond * 100
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	obs.assertErrs(errors.New("future timeout"))
	cli.Close()
	safeReceive(done)
}

func TestMqttClientSubscribe(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := NewSuback()
	suback.ReturnCodes = []QOS{QOSFailure}
	suback.ID = 1

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	err = cli.Subscribe([]Subscription{Subscription{Topic: "test"}})
	assert.NoError(t, err)
	obs.assertErrs(errors.New("failed subscription"))

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientReconnect(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	publish := NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker1 := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Close()

	broker2 := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Close()

	broker3 := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2, broker3)

	cc := newConfig(port)
	cc.Timeout = time.Second
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	cli.Send(publish)
	obs.assertPkts(publish)
	obs.assertErrs(io.EOF)

	cli.Send(publish)
	obs.assertPkts(publish)
	obs.assertErrs(io.EOF)

	cli.Send(publish)
	obs.assertPkts(publish)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientReconnect2(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	publish := NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker1 := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	broker2 := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	broker3 := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2, broker3)

	cc := newConfig(port)
	cc.Timeout = time.Second
	obs := newMockObserver(t)
	cli, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	obs.assertErrs(io.EOF)
	obs.assertErrs(ErrFutureCanceled)
	obs.assertErrs(io.EOF)
	obs.assertErrs(ErrFutureCanceled)

	cli.Send(publish)
	obs.assertPkts(publish)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}
