package link

import (
	"context"
	"errors"
	fmt "fmt"
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
	c, err := NewClient(ClientConfig{}, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	defer c.Close()

	ctx, cel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel()
	req := &Message{}
	res, err := c.CallContext(ctx, req)
	assert.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = latest connection error: connection error: desc = \"transport: Error while dialing dial tcp: missing address\"")
	assert.Nil(t, res)
}

func TestClientConnectErrorWrongPort(t *testing.T) {
	cc := ClientConfig{Address: "localhost:123456789"}
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	defer c.Close()

	ctx, cel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cel()
	req := &Message{}
	res, err := c.CallContext(ctx, req)
	assert.EqualError(t, err, "rpc error: code = DeadlineExceeded desc = latest connection error: connection error: desc = \"transport: Error while dialing dial tcp: address 123456789: invalid port\"")
	assert.Nil(t, res)
}

func TestClientConnectCallSend(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg := &Message{}
	msg.Context.ID = 1
	ack := &Message{}
	ack.Context.ID = 1
	ack.Context.Flags = 0x2

	server := flow.New().Debug().
		Receive(msg).
		Send(msg).
		Receive(ack).
		Send(ack).
		End()

	done := fakeServer(t, server)

	cc := newClientConfig()
	cc.DisableAutoAck = false
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, res)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)
	obs.assertMsgs(ack)
	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientConnectDisableAutoAck(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg := &Message{}
	msg.Context.ID = 1
	ack := &Message{}
	ack.Context.ID = 1
	ack.Context.Flags = 0x2

	server := flow.New().Debug().
		Receive(msg).
		Send(msg).
		Receive(ack).
		Send(ack).
		End()

	done := fakeServer(t, server)

	cc := newClientConfig()
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, res)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)

	err = c.Send(ack)
	assert.NoError(t, err)

	obs.assertMsgs(ack)
	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientConnectWithoutCredentials(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg := &Message{}
	msg.Context.ID = 1

	server := flow.New().Debug().
		Receive(msg).
		Send(msg).
		End()

	done := fakeServer(t, server)

	cc := newClientConfig()
	cc.Username = ""
	cc.Password = ""

	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg)
	assert.EqualError(t, err, "rpc error: code = Unauthenticated desc = Username is unauthenticated")
	assert.Nil(t, res)

	err = c.Send(msg)
	assert.NoError(t, err)

	obs.assertErrs(ErrUnauthenticated)
	c.Close()

	fmt.Println("--> wrong password <--")
	time.Sleep(time.Second)

	cc = newClientConfig()
	cc.Username = "u1"
	cc.Password = "p2"

	c, err = NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.Send(msg)
	assert.NoError(t, err)

	obs.assertErrs(ErrUnauthenticated)
	c.Close()

	fmt.Println("--> signal server to end <--")

	cc = newClientConfig()
	c, err = NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestClientReconnect(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg := &Message{}
	msg.Context.ID = 1
	ack := &Message{}
	ack.Context.ID = 1
	ack.Context.Flags = 0x2

	server := flow.New().Debug().
		Receive(msg).
		Send(msg).
		Close()
	done := fakeServer(t, server)

	cc := newClientConfig()
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, res)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)

	obs.assertErrs(errors.New("rpc error: code = Unavailable desc = transport is closing"))
	safeReceive(done)

	server = flow.New().Debug().
		Receive(msg).
		Send(msg).
		Close()
	done = fakeServer(t, server)

	time.Sleep(time.Second)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)

	obs.assertErrs(errors.New("rpc error: code = Unavailable desc = transport is closing"))
	safeReceive(done)

	assert.NoError(t, c.Close())
}
