package mqtt

import (
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport/flow"
	"github.com/stretchr/testify/assert"
)

func TestClientConnectErrorMissingAddress(t *testing.T) {
	c, err := NewClient(ClientConfig{}, &mockHandler{t: t})
	assert.EqualError(t, err, "parse : empty url")
	assert.Nil(t, c)
}

func TestClientConnectErrorWrongPort(t *testing.T) {
	cc := newConfig("1234567")
	c, err := NewClient(cc, &mockHandler{t: t})
	assert.EqualError(t, err, "dial tcp: address 1234567: invalid port")
	assert.Nil(t, c)
}

func TestClientConnect(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	c, err := NewClient(cc, &mockHandler{t: t})
	assert.NoError(t, err)
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
	c, err := NewClient(cc, &mockHandler{t: t})
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestClientConnectWithCredentials(t *testing.T) {
	connect := connectPacket()
	connect.Username = "test"
	connect.Password = "test"

	connack := connackPacket()
	connack.ReturnCode = packet.BadUsernameOrPassword

	broker := flow.New().Debug().
		Receive(connect).
		Send(connack).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Username = "test"
	cc.Password = "test"
	ch := &mockHandler{t: t, expectedError: "connection refused: bad user name or password"}
	c, err := NewClient(cc, ch)
	assert.EqualError(t, err, "connection refused: bad user name or password")
	assert.Nil(t, c)

	safeReceive(done)
}

func TestClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = packet.NotAuthorized

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	ch := &mockHandler{t: t, expectedError: "connection refused: not authorized"}
	c, err := NewClient(cc, ch)
	assert.Nil(t, c)
	assert.EqualError(t, err, "connection refused: not authorized")

	safeReceive(done)
}

func TestClientExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(packet.NewPingresp()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	ch := &mockHandler{t: t, expectedError: "client expected connack"}
	c, err := NewClient(cc, ch)
	assert.Nil(t, c)
	assert.EqualError(t, err, "client expected connack")

	safeReceive(done)
}

func TestClientNotExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Send(connackPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	ch := &mockHandler{t: t, expectedError: "client already connecting"}
	c, err := NewClient(cc, ch)
	assert.NoError(t, err)

	safeReceive(done)

	assert.NoError(t, c.Close())
}

func TestClientKeepAlive(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreq()
	pingresp := packet.NewPingresp()

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
	c, err := NewClient(cc, &mockHandler{t: t})
	assert.NoError(t, err)

	<-time.After(250 * time.Millisecond)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestClientKeepAliveTimeout(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 0

	pingreq := packet.NewPingreq()

	broker := flow.New().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.KeepAlive = time.Millisecond * 5
	ch := &mockHandler{t: t, expectedError: "client missing pong"}
	c, err := NewClient(cc, ch)
	assert.NoError(t, err)

	safeReceive(done)

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
	c, err := NewClient(cc, &mockHandler{t: t})
	assert.NoError(t, err)

	<-time.After(250 * time.Millisecond)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{0}
	suback.ID = 1

	publish := packet.NewPublish()
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

	wait := make(chan struct{})

	callback := func(p *packet.Publish) error {
		assert.Equal(t, "test", p.Message.Topic)
		assert.Equal(t, []byte("test"), p.Message.Payload)
		assert.Equal(t, packet.QOS(0), p.Message.QOS)
		assert.False(t, p.Message.Retain)
		close(wait)
		return nil
	}
	cc := newConfig(port)
	cc.Subscriptions = []QOSTopic{{Topic: "test"}}
	ch := &mockHandler{t: t, expectedProcessPublish: callback}
	c, err := NewClient(cc, ch)
	assert.NoError(t, err)

	err = c.Send(publish)
	assert.NoError(t, err)

	safeReceive(wait)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestClientPublishSubscribeQOS1(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test", QOS: 1}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{1}
	suback.ID = 1

	publish := packet.NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")
	publish.Message.QOS = 1
	publish.ID = 2

	puback := packet.NewPuback()
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

	wait := make(chan struct{})

	callback1 := func(p *packet.Publish) error {
		assert.Equal(t, "test", p.Message.Topic)
		assert.Equal(t, []byte("test"), p.Message.Payload)
		assert.Equal(t, packet.QOS(1), p.Message.QOS)
		assert.False(t, p.Message.Retain)
		close(wait)
		return nil
	}
	callback2 := func(p *packet.Puback) error {
		assert.Equal(t, packet.ID(2), p.ID)
		return nil
	}
	cc := newConfig(port)
	cc.Subscriptions = []QOSTopic{{Topic: "test", QOS: 1}}
	ch := &mockHandler{t: t, expectedProcessPublish: callback1, expectedProcessPuback: callback2}
	c, err := NewClient(cc, ch)
	assert.NoError(t, err)

	err = c.Send(publish)
	assert.NoError(t, err)

	safeReceive(wait)

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
	ch := &mockHandler{t: t, expectedError: "EOF"}
	c, err := NewClient(cc, ch)
	assert.NoError(t, err)

	safeReceive(done)
	time.Sleep(time.Millisecond * 100)

	assert.NoError(t, c.Close())
}

func TestClientConnackFutureCancellation(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	ch := &mockHandler{t: t, expectedError: "EOF"}
	c, err := NewClient(cc, ch)
	assert.Nil(t, c)
	assert.EqualError(t, err, "EOF")

	safeReceive(done)
}

func TestClientConnackFutureTimeout(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Timeout = time.Millisecond * 50
	ch := &mockHandler{t: t, expectedError: "failed to wait connect ack: future timeout"}
	c, err := NewClient(cc, ch)
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed to wait connect ack: future timeout")

	safeReceive(done)
}

func TestClientSubscribeFutureTimeout(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
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
	ch := &mockHandler{t: t, expectedError: "failed to wait subscribe ack: future timeout"}
	c, err := NewClient(cc, ch)
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed to wait subscribe ack: future timeout")

	safeReceive(done)
}

func TestClientSubscribeValidate(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{packet.QOSFailure}
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
	ch := &mockHandler{t: t, expectedError: "failed subscription"}
	c, err := NewClient(cc, ch)
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed subscription")

	safeReceive(done)
}

func TestClientSubscribeWithoutValidate(t *testing.T) {
	subscribe := packet.NewSubscribe()
	subscribe.Subscriptions = []packet.Subscription{{Topic: "test"}}
	subscribe.ID = 1

	suback := packet.NewSuback()
	suback.ReturnCodes = []packet.QOS{packet.QOSFailure}
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
	c, err := NewClient(cc, &mockHandler{t: t})
	assert.NotNil(t, c)
	assert.NoError(t, err)

	assert.NoError(t, c.Close())

	safeReceive(done)
}
