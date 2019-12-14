package mqtt

import (
	"time"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/jpillora/backoff"
)

// Client auto reconnection client
type Client struct {
	cfg   ClientConfig
	obs   Observer
	cache chan Packet
	log   *log.Logger
	tomb  utils.Tomb
}

// NewClient creates a new client
func NewClient(cc ClientConfig, obs Observer) *Client {
	c := &Client{
		cfg:   cc,
		obs:   obs,
		cache: make(chan Packet, cc.BufferSize),
		log:   log.With(log.Any("cid", cc.ClientID)),
	}
	c.tomb.Go(c.connecting)
	return c
}

// Send sends a generic packet
func (c *Client) Send(pkt Packet) error {
	select {
	case c.cache <- pkt:
	case <-c.tomb.Dying():
		return ErrClientAlreadyClosed
	}
	return nil
}

// Close closes client
func (c *Client) Close() error {
	c.tomb.Kill(nil)
	return c.tomb.Wait()
}

func (c *Client) connecting() error {
	c.log.Info("client starts to connect")
	defer c.log.Info("client has stopped connecting")

	var dying bool
	var current Packet
	bf := backoff.Backoff{
		Min:    time.Second,
		Max:    c.cfg.Interval,
		Factor: 1.6,
	}

	for {
		client, err := newClient(c.cfg, c.obs)
		if err != nil {
			if !c.tomb.Alive() {
				return nil
			}

			c.onError(err)
			c.log.Error("failed to connect", log.Error(err))

			// get backoff duration
			next := bf.Duration()

			c.log.Debug("delay reconnect", log.Any("next", next))

			// sleep but return on Stop
			select {
			case <-time.After(next):
			case <-c.tomb.Dying():
				return nil
			}
			c.log.Debug("next reconnect", log.Any("attempt", bf.Attempt()))
			continue
		}

		bf.Reset()
		c.log.Debug("client online")
		current, dying = c.dispatcher(client, current)
		c.log.Debug("client offline")

		// return goroutine if dying
		if dying {
			return nil
		}
	}
}

// reads from the queues and calls the current client
func (c *Client) dispatcher(cli *client, current Packet) (Packet, bool) {
	defer cli.Close()

	if current != nil {
		err := cli.Send(current)
		if err != nil {
			return current, false
		}
	}

	for {
		select {
		case pkt := <-c.cache:
			err := cli.Send(pkt)
			if err != nil {
				return pkt, false
			}
		case <-c.tomb.Dying():
			return nil, true
		}
	}
}

func (c *Client) onError(err error) {
	if c.obs == nil {
		return
	}
	c.obs.OnError(err)
}
