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

func TestLinkClientConnectErrorWrongPort(t *testing.T) {
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

	done := initMockServer(t, server, nil)

	cc := newClientConfig()
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
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

	done := initMockServer(t, server, nil)

	cc := newClientConfig()
	cc.DisableAutoAck = true
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
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

func TestLinkClientConnectWithoutCredentials(t *testing.T) {
	cfg := log.Config{}
	utils.SetDefaults(&cfg)
	cfg.Level = "debug"
	log.Init(cfg)

	msg := &Message{}
	msg.Context.ID = 1

	server := flow.New().Debug().
		Receive(msg).
		Send(msg).
		Receive(msg).
		End().
		Close()
	done := initMockServer(t, server, &mockAuth{"u1": "p1", "u2": "p2"})

	fmt.Println("--> no password <--")

	cc := newClientConfig()
	cc.Username = ""
	cc.Password = ""

	o1 := newMockObserver(t)
	c1, err := NewClient(cc, o1)
	assert.NoError(t, err)
	assert.NotNil(t, c1)

	res, err := c1.Call(msg)
	assert.EqualError(t, err, "rpc error: code = Unauthenticated desc = Username is unauthenticated")
	assert.Nil(t, res)
	o1.assertErrs(ErrUnauthenticated)
	c1.Close()

	fmt.Println("--> wrong password <--")

	cc = newClientConfig()
	cc.Username = "u1"
	cc.Password = "p2"

	o2 := newMockObserver(t)
	c2, err := NewClient(cc, o2)
	assert.NoError(t, err)
	assert.NotNil(t, c2)

	err = c2.Send(msg)
	assert.NoError(t, err)
	o2.assertErrs(ErrUnauthenticated)
	c2.Close()

	fmt.Println("--> signal server to end <--")

	o3 := newMockObserver(t)
	cc = newClientConfig()
	c3, err := NewClient(cc, o3)
	assert.NoError(t, err)
	assert.NotNil(t, c3)

	err = c3.Send(msg)
	assert.NoError(t, err)
	o3.assertMsgs(msg)
	err = c3.Send(msg)
	assert.NoError(t, err)
	c3.Close()

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
	done := initMockServer(t, server, nil)

	cc := newClientConfig()
	cc.Timeout = time.Millisecond * 100
	obs := newMockObserver(t)
	c, err := NewClient(cc, obs)
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
	done = initMockServer(t, server, nil)

	err = c.Send(msg)
	assert.NoError(t, err)
	obs.assertMsgs(msg)

	assert.NoError(t, c.Close())
	safeReceive(done)
}
