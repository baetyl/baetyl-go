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

func TestClientConnectErrorMissingAddress(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	obs := newMockObserver(t)
	c := NewClient(ClientConfig{}, obs)
	assert.NotNil(t, c)
	defer c.Close()

	obs.assertErrs(errors.New("parse : empty url"))
}

func TestClientConnectErrorWrongPort(t *testing.T) {
	cc := newConfig("1234567")
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)
	defer c.Close()

	obs.assertErrs(errors.New("dial tcp: address 1234567: invalid port"))
}

func TestClientConnect(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)
	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientConnectCustomDialer(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)
	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientConnectWithCredentials(t *testing.T) {
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
	c := NewClient(cc, obs)
	assert.NotNil(t, c)
	defer c.Close()

	obs.assertErrs(errors.New("connection refused: bad user name or password"))
	safeReceive(done)
}

func TestClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = NotAuthorized

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)
	defer c.Close()

	obs.assertErrs(errors.New("connection refused: not authorized"))
	safeReceive(done)
}

func TestClientExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(NewPingresp()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)
	defer c.Close()

	obs.assertErrs(ErrClientExpectedConnack)
	safeReceive(done)
}

func TestClientNotExpectedConnack(t *testing.T) {
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
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	obs.assertErrs(ErrClientAlreadyConnecting)
	safeReceive(done)
	assert.NoError(t, c.Close())
}

func TestClientKeepAlive(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

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
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	<-time.After(250 * time.Millisecond)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientKeepAliveTimeout(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := NewPingreq()

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.KeepAlive = time.Millisecond * 5
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	safeReceive(done)
	obs.assertErrs(ErrClientMissingPong)
	assert.NoError(t, c.Close())
}

func TestClientKeepAliveNone(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.KeepAlive = -1
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	<-time.After(250 * time.Millisecond)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
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
	cc.Subscriptions = []QOSTopic{{Topic: "test"}}
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	err := c.Send(publish)
	assert.NoError(t, err)
	obs.assertPkts(publish)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
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
	cc.Subscriptions = []QOSTopic{{Topic: "test", QOS: 1}}
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	err := c.Send(publish)
	assert.NoError(t, err)

	obs.assertPkts(puback, publish)

	err = c.Send(puback)
	assert.NoError(t, err)
	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientUnexpectedClose(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	safeReceive(done)
	obs.assertErrs(io.EOF)
	assert.NoError(t, c.Close())
}

func TestClientConnackFutureCancellation(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	obs.assertErrs(io.EOF)
	safeReceive(done)
}

func TestClientConnackFutureTimeout(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Timeout = time.Millisecond * 50
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	obs.assertErrs(errors.New("future timeout"))
	safeReceive(done)
}

func TestClientSubscribeFutureTimeout(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test"}}
	subscribe.ID = 1

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Timeout = time.Millisecond * 50
	cc.Subscriptions = []QOSTopic{QOSTopic{Topic: "test"}}
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	obs.assertErrs(errors.New("future timeout"))
	safeReceive(done)
}

func TestClientSubscribeValidate(t *testing.T) {
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
	cc.ValidateSubs = true
	cc.Subscriptions = []QOSTopic{QOSTopic{Topic: "test"}}
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	obs.assertErrs(errors.New("failed subscription"))
	safeReceive(done)
}

func TestClientSubscribeWithoutValidate(t *testing.T) {
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
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Subscriptions = []QOSTopic{QOSTopic{Topic: "test"}}
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientReconnect(t *testing.T) {
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
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker1, broker2)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c := NewClient(cc, obs)
	assert.NotNil(t, c)

	c.Send(publish)
	obs.assertPkts(publish)
	obs.assertErrs(io.EOF)
	c.Send(publish)
	obs.assertPkts(publish)

	assert.NoError(t, c.Close())
	safeReceive(done)
}
