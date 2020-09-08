package mqtt

import (
	"time"

	"github.com/jpillora/backoff"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

// Client auto reconnection client
type Client struct {
	ops   *ClientOptions
	ids   *Counter
	cache chan Packet
	log   *log.Logger
	tomb  utils.Tomb
}

// NewClient creates a new client
func NewClient(ops *ClientOptions) *Client {
	c := &Client{
		ops:   ops,
		ids:   NewCounter(),
		cache: make(chan Packet, ops.MaxCacheMessages),
		log:   log.With(log.Any("mqtt", "client"), log.Any("cid", ops.ClientID)),
	}
	return c
}

func (c *Client) Start(obs Observer) error {
	return c.tomb.Go(func() error {
		return c.connecting(obs)
	})
}

// Publish sends a publish packet
func (c *Client) Publish(qos QOS, topic string, payload []byte, pid ID, retain bool, dup bool) error {
	publish := NewPublish()
	publish.ID = pid
	publish.Dup = dup
	publish.Message.QOS = qos
	publish.Message.Topic = topic
	publish.Message.Payload = payload
	publish.Message.Retain = retain
	if qos != 0 && pid == 0 {
		publish.ID = c.ids.NextID()
	}
	return c.Send(publish)
}

// Send sends a generic packet
func (c *Client) Send(pkt Packet) error {
	select {
	case c.cache <- pkt:
		return nil
	case <-c.tomb.Dying():
		return errors.Trace(ErrClientAlreadyClosed)
	}
}

// Send sends a generic packet, drop the packet if the channel is full
func (c *Client) SendOrDrop(pkt Packet) error {
	select {
	case c.cache <- pkt:
		return nil
	case <-c.tomb.Dying():
		return errors.Trace(ErrClientAlreadyClosed)
	default:
		c.log.Warn("client dropped a packet", log.Any("packet", pkt))
		return nil
	}
}

// Close closes client
func (c *Client) Close() error {
	c.log.Info("client is closing")
	defer c.log.Info("client has closed")

	c.tomb.Kill(nil)
	return errors.Trace(c.tomb.Wait())
}

func (c *Client) connecting(obs Observer) error {
	c.log.Info("client starts to keep connecting")
	defer c.log.Info("client has stopped connecting")

	var err error
	var curr Packet
	var stream *stream
	var next time.Time
	timer := time.NewTimer(0)
	defer timer.Stop()
	bf := backoff.Backoff{
		Min:    time.Second,
		Max:    c.ops.MaxReconnectInterval,
		Factor: 1.6,
	}

	for {
		if !next.IsZero() {
			timer.Reset(next.Sub(time.Now()))
			c.log.Info("next reconnect", log.Any("at", next), log.Any("attempt", bf.Attempt()))
		}
		if stream != nil {
			stream.close()
			stream = nil
			c.log.Info("client has disconnected")
		}
		select {
		case <-c.tomb.Dying():
			return nil
		case <-timer.C:
		}

		c.log.Info("client starts to connect")
		next = time.Now().Add(bf.Duration())
		stream, err = c.connect(obs)
		if err != nil {
			c.log.Error("failed to connect", log.Error(err))
			continue
		}
		c.log.Info("client has connected")
		bf.Reset()
		curr = stream.sending(curr)
	}
}
