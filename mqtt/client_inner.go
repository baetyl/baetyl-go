package mqtt

import (
	"fmt"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

// A client connects to a broker and handles the transmission of packets
type client struct {
	conn            Connection
	config          ClientConfig
	tracker         *Tracker
	connectFuture   *Future
	subscribeFuture *Future
	obs             Observer
	log             *log.Logger
	utils.Tomb
	sync.Once
}

// NewClient returns a new client
func newClient(cc ClientConfig, obs Observer) (*client, error) {
	dialer, err := NewDialer(cc.Certificate, cc.Timeout)
	if err != nil {
		return nil, err
	}
	conn, err := dialer.Dial(cc.Address)
	if err != nil {
		return nil, err
	}
	c := &client{
		conn:            conn,
		config:          cc,
		obs:             obs,
		connectFuture:   NewFuture(),
		subscribeFuture: NewFuture(),
		tracker:         NewTracker(cc.KeepAlive),
		log:             log.With(log.Any("cid", cc.ClientID)),
	}
	err = c.connect()
	if err != nil {
		return nil, c.close(err)
	}
	return c, nil
}

func (c *client) connect() (err error) {
	// allocate packet
	connect := NewConnect()
	connect.ClientID = c.config.ClientID
	connect.KeepAlive = uint16(c.config.KeepAlive.Seconds())
	connect.CleanSession = c.config.CleanSession
	connect.Username = c.config.Username
	connect.Password = c.config.Password
	// connect.Will = c.config.WillMessage

	// send connect packet
	err = c.send(connect, false)
	if err != nil {
		return err
	}

	// start process routine
	c.Go(c.receiving)
	if c.config.KeepAlive > 0 {
		c.Go(c.pinging)
	}

	if len(c.config.Subscriptions) == 0 {
		err = c.connectFuture.Wait(c.config.Timeout)
		if err != nil {
			err = fmt.Errorf("failed to wait connect ack: %s", err.Error())
			c.die(err)
			return err
		}
		return nil
	}

	// allocate subscribe packet
	subscribe := &Subscribe{
		ID:            1,
		Subscriptions: make([]Subscription, 0),
	}
	for _, s := range c.config.Subscriptions {
		subscribe.Subscriptions = append(subscribe.Subscriptions, Subscription{
			Topic: s.Topic,
			QOS:   QOS(s.QOS),
		})
	}

	// send packet
	err = c.send(subscribe, false)
	if err != nil {
		return err
	}

	err = c.connectFuture.Wait(c.config.Timeout)
	if err != nil {
		err = fmt.Errorf("failed to wait connect ack: %s", err.Error())
		c.die(err)
		return err
	}
	c.log.Debug("client is connected")

	err = c.subscribeFuture.Wait(c.config.Timeout)
	if err != nil {
		err = fmt.Errorf("failed to wait subscribe ack: %s", err.Error())
		c.die(err)
		return err
	}
	c.log.Debug("topics are subscribed")
	return nil
}

// Send sends a generic packet
func (c *client) Send(p Packet) (err error) {
	err = c.send(p, true)
	if err != nil {
		c.die(err)
	}
	return
}

// Close closes client
func (c *client) Close() error {
	c.log.Info("client is closing")
	defer c.log.Info("client has closed")
	c.close(nil)
	return nil
}

// closes client by itself
func (c *client) die(err error) {
	if !c.Alive() {
		return
	}
	go func() {
		c.log.Info("client is closing by itself", log.Error(err))
		c.close(err)
		c.log.Info("client has closed by itself")
	}()
}

func (c *client) close(err error) error {
	c.Do(func() {
		c.Kill(err)
		c.connectFuture.Cancel()
		c.subscribeFuture.Cancel()
		if err == nil {
			c.send(NewDisconnect(), false)
		} else {
			c.onError(err)
		}
		if c.conn != nil {
			c.conn.Close()
		}
	})
	return c.Wait()
}

func (c *client) receiving() error {
	c.log.Info("client starts to receive packets")
	defer c.log.Info("client has stopped receiving packets")

	pkt, err := c.conn.Receive()
	if err != nil {
		c.die(err)
		return err
	}
	if ent := c.log.Check(log.DebugLevel, "client received a packet"); ent != nil {
		ent.Write(log.Any("packet", pkt.String()))
	}
	p, ok := pkt.(*Connack)
	if !ok {
		c.die(ErrClientExpectedConnack)
		return ErrClientExpectedConnack
	}
	if p.ReturnCode != ConnectionAccepted {
		err = fmt.Errorf(p.ReturnCode.String())
		c.die(err)
		return err
	}

	c.connectFuture.Complete()

	for {
		// get next packet from connection
		pkt, err := c.conn.Receive()
		if err != nil {
			c.die(err)
			return err
		}

		if ent := c.log.Check(log.DebugLevel, "client received a packet"); ent != nil {
			ent.Write(log.Any("packet", pkt.String()))
		}

		switch p := pkt.(type) {
		case *Publish:
			err = c.onPublish(p)
		case *Puback:
			err = c.onPuback(p)
		case *Suback:
			if c.config.ValidateSubs {
				for _, code := range p.ReturnCodes {
					if code == QOSFailure {
						err = ErrFailedSubscription
						c.die(err)
						return err
					}
				}
			}
			c.subscribeFuture.Complete()
		case *Pingresp:
			c.tracker.Pong()
		case *Connack:
			err = ErrClientAlreadyConnecting
		default:
			err = fmt.Errorf("packet (%v) not supported", p)
		}

		if err != nil {
			c.die(err)
			return err
		}
	}
}

func (c *client) pinging() (err error) {
	c.log.Info("client starts to send ping")
	defer c.log.Info("client has stopped sending ping")

	for {
		// get current window
		window := c.tracker.Window()

		// check if ping is due
		if window < 0 {
			// check if a pong has already been sent
			if c.tracker.Pending() {
				c.die(ErrClientMissingPong)
				return ErrClientMissingPong
			}

			// send pingreq packet
			err = c.send(NewPingreq(), false)
			if err != nil {
				c.die(err)
				return err
			}

			// save ping attempt
			c.tracker.Ping()
		}

		select {
		case <-c.Dying():
			return nil
		case <-time.After(window):
			continue
		}
	}
}

/* helpers */

// sends packet and updates lastSend
func (c *client) send(pkt Packet, async bool) error {

	// reset keep alive tracker
	c.tracker.Reset()

	// send packet
	err := c.conn.Send(pkt, async)
	if err != nil {
		c.die(err)
		return err
	}

	if ent := c.log.Check(log.DebugLevel, "client sent a packet"); ent != nil {
		ent.Write(log.Any("packet", pkt.String()))
	}

	return nil
}

func (c *client) onPublish(pkt *Publish) error {
	if c.obs == nil {
		return nil
	}
	return c.obs.OnPublish(pkt)
}

func (c *client) onPuback(pkt *Puback) error {
	if c.obs == nil {
		return nil
	}
	return c.obs.OnPuback(pkt)
}

func (c *client) onError(err error) {
	if c.obs == nil {
		return
	}
	c.obs.OnError(err)
}
