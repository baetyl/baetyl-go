package mqtt

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mock"
	"github.com/baetyl/baetyl-go/v2/utils"
)

func TestMqttClientConnectErrorMissingAddress(t *testing.T) {
	ops := newClientOptions(t, "", nil)
	ops.Address = ""
	cli := NewClient(ops)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)
}

func TestMqttClientConnectErrorWrongPort(t *testing.T) {
	ops := newClientOptions(t, "1234567", nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)
}

func TestMqttClientConnectWithCredentials(t *testing.T) {
	connect := connectPacket()
	connect.Username = "test"
	connect.Password = "test"

	connack := connackPacket()
	connack.ReturnCode = BadUsernameOrPassword

	broker := mock.NewFlow().Debug().
		Receive(connect).
		Send(connack).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	ops.Username = "test"
	ops.Password = "test"
	cli := NewClient(ops)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)
	obs.assertErrs(errors.New("connection refused: bad user name or password"))
	safeReceive(done)
}

func TestMqttClientConnectionDenied(t *testing.T) {
	connack := connackPacket()
	connack.ReturnCode = NotAuthorized

	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connack).
		Close()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)
	obs.assertErrs(errors.New("connection refused: not authorized"))
	safeReceive(done)
}

func TestMqttClientExpectedConnack(t *testing.T) {
	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(NewPingresp()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)
	defer cli.Close()

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)
	obs.assertErrs(ErrClientExpectedConnack)
	safeReceive(done)
}

func TestMqttClientNotExpectedConnack(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Send(connackPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)
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
	broker := mock.NewFlow().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(pingreq).
		Send(pingresp).
		Receive(pingreq).
		Send(pingresp).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	ops.KeepAlive = time.Millisecond * 100
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	time.Sleep(250 * time.Millisecond)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientKeepAliveTimeout(t *testing.T) {
	connect := connectPacket()
	connect.KeepAlive = 1

	broker := mock.NewFlow().Debug().
		Receive(connect).
		Send(connackPacket()).
		Receive(NewPingreq()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	ops.KeepAlive = time.Millisecond * 100
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	safeReceive(done)
	obs.assertErrs(ErrClientMissingPong)
	assert.NoError(t, cli.Close())
}

func TestMqttClientKeepAliveNone(t *testing.T) {
	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	<-time.After(time.Second)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientPublishSubscribeQOS0(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test"}}
	subscribe.ID = subscribeId

	suback := NewSuback()
	suback.ReturnCodes = []QOS{0}
	suback.ID = subscribeId

	publish := NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, []Subscription{{Topic: "test"}})
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	err = cli.Publish(publish.Message.QOS, publish.Message.Topic, publish.Message.Payload, publish.ID, publish.Message.Retain, publish.Dup)
	assert.NoError(t, err)
	obs.assertPkts(publish)

	time.Sleep(time.Second)

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

	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(publish).
		Send(puback).
		Send(publish).
		Receive(puback).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, []Subscription{{Topic: "test", QOS: 1}})
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	err = cli.Publish(publish.Message.QOS, publish.Message.Topic, publish.Message.Payload, publish.ID, publish.Message.Retain, publish.Dup)
	assert.NoError(t, err)

	obs.assertPkts(puback, publish)

	err = cli.Send(puback)
	assert.NoError(t, err)

	obs.assertPkts(publish)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientAutoAck(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test", QOS: 1}}
	subscribe.ID = 1

	suback := NewSuback()
	suback.ReturnCodes = []QOS{1}
	suback.ID = 1

	pub0 := NewPublish()
	pub0.Message.Topic = "test"
	pub0.Message.Payload = []byte("test")

	pub1 := NewPublish()
	pub1.Message.Topic = "test"
	pub1.Message.Payload = []byte("test")
	pub1.Message.QOS = 1
	pub1.ID = 2

	puback := NewPuback()
	puback.ID = 2

	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		Receive(pub1).
		Send(puback).
		Send(pub1).
		Receive(puback). // auto ack
		Send(pub0).
		Receive(puback).
		Send(pub1). // not auto ack since user code error
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, []Subscription{{Topic: "test", QOS: 1}})
	ops.DisableAutoAck = false
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	err = cli.Publish(pub1.Message.QOS, pub1.Message.Topic, pub1.Message.Payload, pub1.ID, pub1.Message.Retain, pub1.Dup)
	assert.NoError(t, err)

	obs.assertPkts(puback, pub1, pub0)

	obs.setErrOnPublish(ErrFutureTimeout)
	err = cli.Send(puback)
	assert.NoError(t, err)

	obs.assertPkts(pub1)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientUnexpectedClose(t *testing.T) {
	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Close()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	safeReceive(done)
	obs.assertErrs(io.EOF)
	assert.NoError(t, cli.Close())
}

func TestMqttClientConnackTimeout1(t *testing.T) {
	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Close()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	obs.assertErrs(io.EOF)
	safeReceive(done)
	cli.Close()
}

func TestMqttClientConnackTimeout2(t *testing.T) {
	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, nil)
	ops.Timeout = time.Millisecond * 100
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	obs.assertErrs(ErrFutureTimeout)
	cli.Close()
	safeReceive(done)
}

func TestMqttClientSubscribe(t *testing.T) {
	subscribe := NewSubscribe()
	subscribe.Subscriptions = []Subscription{{Topic: "test"}}
	subscribe.ID = subscribeId

	suback := NewSuback()
	suback.ReturnCodes = []QOS{QOSFailure}
	suback.ID = subscribeId

	broker := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(subscribe).
		Send(suback).
		End()

	done, port := initMockBroker(t, broker)

	ops := newClientOptions(t, port, []Subscription{{Topic: "test"}})
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	time.Sleep(time.Second)

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

	broker1 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Close()

	broker2 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Close()

	broker3 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker1, broker2, broker3)

	ops := newClientOptions(t, port, nil)
	ops.Timeout = time.Second
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

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

	broker1 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Close()

	broker2 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Close()

	broker3 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker1, broker2, broker3)

	ops := newClientOptions(t, port, nil)
	ops.Timeout = time.Second
	cli := NewClient(ops)
	assert.NotNil(t, cli)

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	obs.assertErrs(io.EOF)
	obs.assertErrs(io.EOF)

	cli.Send(publish)
	obs.assertPkts(publish)

	assert.NoError(t, cli.Close())
	safeReceive(done)
}

func TestMqttClientReconnect3(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	publish := NewPublish()
	publish.Message.Topic = "test"
	publish.Message.Payload = []byte("test")

	connect := connectPacket()
	connect.Username = "test"
	connect.Password = "test"
	connect.ClientID = "test"

	connack := connackPacket()
	connack.ReturnCode = BadUsernameOrPassword

	broker1 := mock.NewFlow().Debug().
		Receive(connect).
		Send(connack).
		End()

	broker2 := mock.NewFlow().Debug().
		Receive(connectPacket()).
		Send(connackPacket()).
		Receive(publish).
		Send(publish).
		Receive(disconnectPacket()).
		End()

	done, port := initMockBroker(t, broker1, broker2)

	ops := newClientOptions(t, port, nil)
	ops.Timeout = time.Second
	ops.ClientID = "test"
	ops.Username = "test"
	ops.Password = "test"
	cli := NewClient(ops)
	assert.NotNil(t, cli)
	cli.SetReconnectCallback(func() error {
		cli.ResetClient(&ClientOptions{
			ClientID: "",
			Username: "",
			Password: "",
		})
		return nil
	})

	obs := newMockObserver(t)
	err := cli.Start(obs)
	assert.NoError(t, err)

	obs.assertErrs(errors.New("connection refused: bad user name or password"))

	cli.Send(publish)
	obs.assertPkts(publish)

	assert.Equal(t, cli.ops.ClientID, "")
	assert.Equal(t, cli.ops.Username, "")
	assert.Equal(t, cli.ops.Password, "")

	assert.NoError(t, cli.Close())
	safeReceive(done)
}
