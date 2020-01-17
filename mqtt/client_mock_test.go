package mqtt

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl-go/flow"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
)

type mockObserver struct {
	t            *testing.T
	pkts         chan Packet
	errs         chan error
	errOnPublish error
	sync.Mutex
}

func newMockObserver(t *testing.T) *mockObserver {
	return &mockObserver{
		t:    t,
		pkts: make(chan Packet, 10),
		errs: make(chan error, 10),
	}
}

func (o *mockObserver) OnPublish(pkt *packet.Publish) error {
	fmt.Println("--> OnPublish:", pkt)
	o.pkts <- pkt
	o.Lock()
	defer o.Unlock()
	return o.errOnPublish
}

func (o *mockObserver) OnPuback(pkt *packet.Puback) error {
	fmt.Println("--> OnPuback:", pkt)
	o.pkts <- pkt
	return nil
}

func (o *mockObserver) OnError(err error) {
	fmt.Println("--> OnError:", err)
	o.errs <- err
}

func (o *mockObserver) setErrOnPublish(err error) {
	o.Lock()
	o.errOnPublish = err
	o.Unlock()
}

func (o *mockObserver) assertPkts(pkts ...Packet) {
	for _, pkt := range pkts {
		select {
		case <-time.After(6 * time.Minute):
			panic("nothing received")
		case p := <-o.pkts:
			assert.Equal(o.t, pkt, p)
		}
	}
}

func (o *mockObserver) assertErrs(errs ...error) {
	for _, err := range errs {
		select {
		case <-time.After(6 * time.Second):
			panic("nothing received")
		case e := <-o.errs:
			assert.Equal(o.t, err.Error(), e.Error())
		}
	}
}

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(1 * time.Second):
		panic("nothing received")
	case <-ch:
	}
}

func newConfig(port string) (c ClientConfig) {
	c.Address = "tcp://localhost:" + port
	c.CleanSession = true
	c.DisableAutoAck = true
	defaults.Set(&c)
	return
}

func initMockBroker(t *testing.T, testFlows ...*flow.Flow) (chan struct{}, string) {
	done := make(chan struct{})

	server, err := transport.Launch("tcp://localhost:0")
	assert.NoError(t, err)

	go func() {
		for _, flow := range testFlows {
			conn, err := server.Accept()
			assert.NoError(t, err)

			err = flow.Test(newWrapper(conn))
			assert.NoError(t, err)
		}

		err = server.Close()
		assert.NoError(t, err)

		close(done)
	}()

	_, port, _ := net.SplitHostPort(server.Addr().String())
	return done, port
}

type wrapper struct {
	Connection
}

func newWrapper(conn Connection) flow.Conn {
	return &wrapper{Connection: conn}
}

func (c *wrapper) Send(pkt interface{}) error {
	return c.Connection.Send(pkt.(Packet), false)
}

func (c *wrapper) Receive() (interface{}, error) {
	pkt, err := c.Connection.Receive()
	if err != nil {
		return nil, err
	}
	return pkt.(Packet), nil
}

func connectPacket() *packet.Connect {
	pkt := packet.NewConnect()
	pkt.CleanSession = true
	return pkt
}

func connackPacket() *packet.Connack {
	pkt := packet.NewConnack()
	pkt.ReturnCode = packet.ConnectionAccepted
	pkt.SessionPresent = false
	return pkt
}

func disconnectPacket() *packet.Disconnect {
	return packet.NewDisconnect()
}
