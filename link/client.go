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
	c.log.Info("client starts to connect")
	defer c.log.Info("client has stopped connecting")

	var dying bool
	var current *Message
	bf := backoff.Backoff{
		Min:    time.Second,
		Max:    c.cfg.Interval,
		Factor: 1.6,
	}

	for {
		ts := time.Now().UnixNano()
		ctx, cel := context.WithTimeout(context.Background(), bf.Duration())
		defer cel()
		s, err := newStream(ctx, c)
		if err != nil {
			if !c.tomb.Alive() {
				return nil
			}
			c.log.Info("next reconnect", log.Any("ts", ts), log.Any("attempt", bf.Attempt()), log.Error(err))
			continue
		}

		bf.Reset()
		c.log.Debug("stream online", log.Any("ts", ts))
		current, dying = c.dispatcher(s, current)
		c.log.Debug("stream offline", log.Any("ts", ts))

		// return goroutine if dying
		if dying {
			return nil
		}
	}
}

// reads from the queues and calls the current client
func (c *Client) dispatcher(s *stream, current *Message) (*Message, bool) {
	defer s.Close()

	if current != nil {
		err := s.Send(current)
		if err != nil {
			return current, false
		}
	}

	for {
		select {
		case msg := <-c.cache:
			err := s.Send(msg)
			if err != nil {
				return msg, false
			}
		case <-s.Dying():
			return nil, false
		case <-c.tomb.Dying():
			return nil, true
		}
	}
}
