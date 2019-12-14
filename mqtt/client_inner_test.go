package mqtt

import (
	"io"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl-go/flow"
	"github.com/stretchr/testify/assert"
)

func TestInnerClientConnectErrorMissingAddress(t *testing.T) {
	c, err := newClient(ClientConfig{}, newMockObserver(t))
	assert.EqualError(t, err, "parse : empty url")
	assert.Nil(t, c)
}

func TestInnerClientConnectErrorWrongPort(t *testing.T) {
	cc := newConfig("1234567")
	c, err := newClient(cc, newMockObserver(t))
	assert.EqualError(t, err, "dial tcp: address 1234567: invalid port")
	assert.Nil(t, c)
}

func TestInnerClientConnect(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	c, err := newClient(cc, newMockObserver(t))
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestInnerClientConnectCustomDialer(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	c, err := newClient(cc, newMockObserver(t))
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestInnerClientConnectWithCredentials(t *testing.T) {
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
	c, err := newClient(cc, newMockObserver(t))
	assert.EqualError(t, err, "connection refused: bad user name or password")
	assert.Nil(t, c)

	safeReceive(done)
}

func TestInnerClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = packet.NotAuthorized

	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	c, err := newClient(cc, newMockObserver(t))
	assert.Nil(t, c)
	assert.EqualError(t, err, "connection refused: not authorized")

	safeReceive(done)
}

func TestInnerClientExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(packet.NewPingresp()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	c, err := newClient(cc, newMockObserver(t))
	assert.Nil(t, c)
	assert.EqualError(t, err, "client expected connack")

	safeReceive(done)
}

func TestInnerClientNotExpectedConnack(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Send(connackPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	obs := newMockObserver(t)
	c, err := newClient(cc, obs)
	assert.NoError(t, err)

	obs.assertErrs(ErrClientAlreadyConnecting)
	safeReceive(done)
	assert.NoError(t, c.Close())
}

func TestInnerClientKeepAlive(t *testing.T) {
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
	c, err := newClient(cc, newMockObserver(t))
	assert.NoError(t, err)

	<-time.After(250 * time.Millisecond)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestInnerClientKeepAliveTimeout(t *testing.T) {
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
	ch := newMockObserver(t)
	c, err := newClient(cc, ch)
	assert.NoError(t, err)

	safeReceive(done)
	ch.assertErrs(ErrClientMissingPong)

	assert.NoError(t, c.Close())
}

func TestInnerClientKeepAliveNone(t *testing.T) {
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
	c, err := newClient(cc, newMockObserver(t))
	assert.NoError(t, err)

	<-time.After(250 * time.Millisecond)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestInnerClientPublishSubscribeQOS0(t *testing.T) {
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

	cc := newConfig(port)
	cc.Subscriptions = []QOSTopic{{Topic: "test"}}
	ch := newMockObserver(t)
	c, err := newClient(cc, ch)
	assert.NoError(t, err)

	err = c.Send(publish)
	assert.NoError(t, err)

	ch.assertPkts(publish)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestInnerClientPublishSubscribeQOS1(t *testing.T) {
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

	cc := newConfig(port)
	cc.Subscriptions = []QOSTopic{{Topic: "test", QOS: 1}}
	ch := newMockObserver(t)
	c, err := newClient(cc, ch)
	assert.NoError(t, err)

	err = c.Send(publish)
	assert.NoError(t, err)

	ch.assertPkts(puback, publish)

	err = c.Send(puback)
	assert.NoError(t, err)

	assert.NoError(t, c.Close())

	safeReceive(done)
}

func TestInnerClientUnexpectedClose(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	ch := newMockObserver(t)
	c, err := newClient(cc, ch)
	assert.NoError(t, err)

	safeReceive(done)

	ch.assertErrs(io.EOF)

	assert.NoError(t, c.Close())
}

func TestInnerClientConnackFutureCancellation(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		Close()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	c, err := newClient(cc, newMockObserver(t))
	assert.Nil(t, c)
	assert.EqualError(t, err, "EOF")

	safeReceive(done)
}

func TestInnerClientConnackFutureTimeout(t *testing.T) {
	broker := flow.New().Debug().
		Receive(connectPacket()).
		End()

	done, port := fakeBroker(t, broker)

	cc := newConfig(port)
	cc.Timeout = time.Millisecond * 50
	c, err := newClient(cc, newMockObserver(t))
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed to wait connect ack: future timeout")

	safeReceive(done)
}

func TestInnerClientSubscribeFutureTimeout(t *testing.T) {
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
	c, err := newClient(cc, newMockObserver(t))
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed to wait subscribe ack: future timeout")

	safeReceive(done)
}

func TestInnerClientSubscribeValidate(t *testing.T) {
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
	c, err := newClient(cc, newMockObserver(t))
	assert.Nil(t, c)
	assert.EqualError(t, err, "failed subscription")

	safeReceive(done)
}

func TestInnerClientSubscribeWithoutValidate(t *testing.T) {
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
	c, err := newClient(cc, newMockObserver(t))
	assert.NotNil(t, c)
	assert.NoError(t, err)

	assert.NoError(t, c.Close())

	safeReceive(done)
}
