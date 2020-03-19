package link

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/jpillora/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ErrClientAlreadyClosed the client is closed
var ErrClientAlreadyClosed = errors.New("client has closed")

// ErrClientMessageTypeInvalid the message type is invalid
var ErrClientMessageTypeInvalid = errors.New("message type is invalid")

// Client client of contact server
type Client struct {
	ops   ClientOptions
	cli   LinkClient
	conn  *grpc.ClientConn
	cache chan *Message
	log   *log.Logger
	tomb  utils.Tomb
}

// NewClient creates a new client of functions server
func NewClient(ops ClientOptions) (*Client, error) {
	// remove link scheme from address if exists
	if addr, err := url.Parse(ops.Address); err == nil && addr.Scheme == "link" {
		ops.Address = addr.Host
	}
	gops := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(ops.MaxMessageSize))),
	}
	if ops.TLSConfig == nil {
		gops = append(gops, grpc.WithInsecure())
	} else {
		gops = append(gops, grpc.WithTransportCredentials(credentials.NewTLS(ops.TLSConfig)))
	}
	conn, err := grpc.Dial(ops.Address, gops...)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		ops:   ops,
		conn:  conn,
		cli:   NewLinkClient(conn),
		cache: make(chan *Message, ops.MaxCacheMessages),
		log:   log.With(log.Any("link", "client")),
	}
	cli.tomb.Go(cli.connecting)
	return cli, nil
}

// Call calls a request synchronously
func (c *Client) Call(msg *Message) (*Message, error) {
	return c.cli.Call(context.Background(), msg, grpc.WaitForReady(true))
}

// CallContext calls a request with context synchronously
func (c *Client) CallContext(ctx context.Context, msg *Message) (*Message, error) {
	return c.cli.Call(ctx, msg, grpc.WaitForReady(true))
}

// Send sends a message asynchronously
func (c *Client) Send(msg *Message) error {
	select {
	case c.cache <- msg:
	case <-c.tomb.Dying():
		return ErrClientAlreadyClosed
	}
	return nil
}

// SendContext sends a message with context asynchronously
func (c *Client) SendContext(ctx context.Context, msg *Message) error {
	select {
	case c.cache <- msg:
	case <-ctx.Done():
		return ctx.Err()
	case <-c.tomb.Dying():
		return ErrClientAlreadyClosed
	}
	return nil
}

// Close closes client
func (c *Client) Close() error {
	c.log.Info("client is closing")
	defer c.log.Info("client has closed")

	c.tomb.Kill(nil)
	err := c.tomb.Wait()
	c.conn.Close()
	return err
}

func (c *Client) connecting() error {
	c.log.Info("client starts to keep connect")
	defer c.log.Info("client has stopped connecting")

	var err error
	var curr *Message
	var next time.Time
	var stream *stream
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
		stream, err = c.connect()
		if err != nil {
			c.onErr("failed to connect", err)
			continue
		}
		c.log.Info("client has connected")
		bf.Reset()
		curr = stream.sending(curr)
	}
}

func (c *Client) onMsg(msg *Message) error {
	if c.ops.Observer == nil {
		return nil
	}
	return c.ops.Observer.OnMsg(msg)
}

func (c *Client) onAck(msg *Message) error {
	if c.ops.Observer == nil {
		return nil
	}
	return c.ops.Observer.OnAck(msg)
}

func (c *Client) onErr(msg string, err error) {
	if c.ops.Observer == nil || err == nil {
		return
	}
	c.log.Error(msg, log.Error(err))
	c.ops.Observer.OnErr(err)
}
