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

func TestLinkClientConnectErrorMissingAddress(t *testing.T) {
	ops := newClientOptions(t)
	ops.Address = ""
	c, err := NewClient(ops)
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

func TestLinkClientConnectErrorWrongPort(t *testing.T) {
	ops := newClientOptions(t)
	ops.Address = "localhost:123456789"
	c, err := NewClient(ops)
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

func TestLinkClientSendRecvMessage(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg0 := &Message{}
	msg1 := &Message{}
	msg1.Context.ID = 1
	msg1.Context.QOS = 1
	ack := &Message{}
	ack.Context.ID = 1
	ack.Context.Type = Ack

	server := flow.New().Debug().
		Receive(msg0).
		Send(msg0).
		Receive(msg1).
		Send(ack).
		Send(msg1).
		Receive(ack). // auto ack
		Receive(ack).
		Send(msg1). // not auto ack since user code error
		End().
		Close()

	done := initMockServer(t, server)

	ops := newClientOptions(t)
	obs := ops.Observer.(*mockObserver)
	c, err := NewClient(ops)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg0)
	assert.NoError(t, err)
	assert.Equal(t, msg0, res)

	res, err = c.Call(msg1)
	assert.NoError(t, err)
	assert.Equal(t, msg1, res)

	err = c.Send(msg0)
	assert.NoError(t, err)
	obs.assertMsgs(msg0)

	err = c.Send(msg1)
	assert.NoError(t, err)
	obs.assertMsgs(ack, msg1)

	obs.setErrOnMsg(ErrClientMessageTypeInvalid)
	err = c.Send(ack)
	assert.NoError(t, err)
	obs.assertMsgs(msg1)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestLinkClientSendRecvMessageDisableAutoAck(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg0 := &Message{}
	msg1 := &Message{}
	msg1.Context.ID = 1
	msg1.Context.QOS = 1
	ack := &Message{}
	ack.Context.ID = 1
	ack.Context.Type = Ack

	server := flow.New().Debug().
		Receive(msg0).
		Send(msg0).
		Receive(msg1).
		Send(ack).
		Send(msg1).
		Receive(ack).
		End().
		Close()

	done := initMockServer(t, server)

	ops := newClientOptions(t)
	ops.DisableAutoAck = true
	ops.Address = "link://" + ops.Address
	obs := ops.Observer.(*mockObserver)
	c, err := NewClient(ops)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg0)
	assert.NoError(t, err)
	assert.Equal(t, msg0, res)

	res, err = c.Call(msg1)
	assert.NoError(t, err)
	assert.Equal(t, msg1, res)

	err = c.Send(msg0)
	assert.NoError(t, err)
	obs.assertMsgs(msg0)

	err = c.Send(msg1)
	assert.NoError(t, err)
	obs.assertMsgs(ack, msg1)

	err = c.Send(ack)
	assert.NoError(t, err)

	assert.NoError(t, c.Close())
	safeReceive(done)
}

func TestLinkClientReconnect(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg := &Message{}
	msg.Context.ID = 1
	ack := &Message{}
	ack.Context.ID = 1
	ack.Context.Type = Ack

	server := flow.New().Debug().
		Receive(msg).
		Close()
	done := initMockServer(t, server)

	ops := newClientOptions(t)
	ops.DisableAutoAck = true
	obs := ops.Observer.(*mockObserver)
	c, err := NewClient(ops)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	res, err := c.Call(msg)
	assert.NoError(t, err)
	assert.Equal(t, msg, res)

	err = c.Send(msg)
	assert.NoError(t, err)

	fmt.Println("--> wait error <--")

	obs.assertErrs(errors.New("rpc error: code = Unavailable desc = transport is closing"))

	fmt.Println("--> wait server close <--")

	safeReceive(done)

	fmt.Println("--> start server again <--")

	server = flow.New().Debug().
		Receive(msg).
		Send(msg).
		End().
		Close()
	done = initMockServer(t, server)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)

	assert.NoError(t, c.Close())
	safeReceive(done)
}
