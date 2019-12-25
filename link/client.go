package link

import (
	"context"
	"errors"
	"time"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/jpillora/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ErrClientAlreadyClosed client is closed
var ErrClientAlreadyClosed = errors.New("client is closed")

// Client client of contact server
type Client struct {
	cfg   ClientConfig
	cli   LinkClient
	obs   Observer
	conn  *grpc.ClientConn
	cache chan *Message
	log   *log.Logger
	tomb  utils.Tomb
}

// NewClient creates a new client of functions server
func NewClient(cc ClientConfig, obs Observer) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cc.MaxMessageSize))),
	}
	// enable tls
	if cc.Certificate.Key != "" || cc.Certificate.Cert != "" {
		tlsCfg, err := utils.NewTLSConfigClient(cc.Certificate)
		if err != nil {
			return nil, err
		}
		if tlsCfg != nil {
			if !cc.InsecureSkipVerify {
				tlsCfg.ServerName = cc.Name
			}
			creds := credentials.NewTLS(tlsCfg)
			opts = append(opts, grpc.WithTransportCredentials(creds))
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	//  enable username/password
	if cc.Username != "" || cc.Password != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(MD{
			KeyUsername: cc.Username,
			KeyPassword: cc.Password,
		}))
	}

	conn, err := grpc.Dial(cc.Address, opts...)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		cfg:   cc,
		obs:   obs,
		conn:  conn,
		cli:   NewLinkClient(conn),
		cache: make(chan *Message, cc.MaxCacheMessages),
		log:   log.With(log.Any("link", "client")),
	}
	cli.tomb.Go(cli.connecting)
	return cli, nil
}

// Call calls a request
func (c *Client) Call(msg *Message) (*Message, error) {
	return c.cli.Call(context.Background(), msg, grpc.WaitForReady(true))
}

// CallContext calls a request with context
func (c *Client) CallContext(ctx context.Context, msg *Message) (*Message, error) {
	return c.cli.Call(ctx, msg, grpc.WaitForReady(true))
}

// Send sends a generic packet
func (c *Client) Send(msg *Message) error {
	select {
	case c.cache <- msg:
	case <-c.tomb.Dying():
		return ErrClientAlreadyClosed
	}
	return nil
}

// SendContext sends a message with context
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
		Max:    c.cfg.Interval,
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
	if c.obs == nil {
		return nil
	}
	return c.obs.OnMsg(msg)
}

func (c *Client) onAck(msg *Message) error {
	if c.obs == nil {
		return nil
	}
	return c.obs.OnAck(msg)
}

func (c *Client) onErr(msg string, err error) {
	if c.obs == nil || err == nil {
		return
	}
	c.log.Error(msg, log.Error(err))
	c.obs.OnErr(err)
}
